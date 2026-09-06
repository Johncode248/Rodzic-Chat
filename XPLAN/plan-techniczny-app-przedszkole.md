# Plan techniczny — aplikacja komunikacji przedszkole–rodzic

Stack: Go (backend/API) + SvelteKit (PWA) + docelowo KMP (iOS/Android), architektura multi-tenant (jedna instalacja obsługuje wiele placówek).

---

## 1. Zakres funkcjonalny MVP

1. Komunikacja z rodzicami (ogłoszenia, czat, powiadomienia push)
2. Zgłaszanie nieobecności dziecka
3. Zapisy na wydarzenia z potwierdzeniem (zgody, np. "wycieczka", "przynieś kasztany")
4. Płatności online (posiłki/zajęcia)
5. Informacyjny plan dnia grupy

---

## 2. Wymagane integracje / API zewnętrzne

### 2.1 Płatności
| Usługa | Do czego | Uwagi |
|---|---|---|
| **Przelewy24** lub **Tpay** | Płatności online (BLIK, karta, przelew) | Wybierz jedną jako główną na start — obie mają REST API + webhooki potwierdzające płatność. Tpay ma prostszy onboarding dla małych firm. |
| Webhook endpoint (Twoje API) | Odbiór potwierdzeń transakcji | Musi być publiczny, weryfikowany podpisem (HMAC) od dostawcy |

### 2.2 Powiadomienia push
| Usługa | Do czego |
|---|---|
| **Firebase Cloud Messaging (FCM)** | Push do Android + iOS (przez APNs pod spodem) + web push dla PWA |
| Web Push API (VAPID) | Jeśli chcesz push w PWA bez pełnego SDK Firebase |

### 2.3 E-mail
| Usługa | Do czego |
|---|---|
| **Resend** lub **SendGrid** | Powiadomienia mailowe (reset hasła, podsumowania, alerty), transakcyjne maile o płatnościach |

### 2.4 SMS (opcjonalnie, faza 2)
| Usługa | Do czego |
|---|---|
| SMSAPI.pl / Twilio | Krytyczne alerty (np. odbiór dziecka), jeśli push zawiedzie |

### 2.5 Uwierzytelnianie
| Opcja | Uwagi |
|---|---|
| Własna implementacja (JWT + refresh token) | Pełna kontrola, mniej zależności zewnętrznych — polecane przy Twoim stacku Go |
| Ewentualnie OAuth (Google) jako dodatkowa opcja logowania | Nie wymagane w MVP |

### 2.6 Przechowywanie plików
| Usługa | Do czego |
|---|---|
| **S3-compatible storage** (Cloudflare R2 / AWS S3 / Backblaze B2) | Zdjęcia z wydarzeń, załączniki do ogłoszeń, avatar dziecka/placówki |

### 2.7 Inne
| Usługa | Do czego |
|---|---|
| Sentry (lub self-hosted) | Monitoring błędów backendu i frontendu |
| Plausible / self-hosted analytics | Analityka bez cookies (RODO-friendly) |

---

## 3. Architektura multi-tenant

Rekomendacja: **jedna baza, jeden schemat, `organization_id` jako kolumna w każdej tabeli** (row-level multi-tenancy) — prostsze w utrzymaniu niż osobne schematy/bazy per placówka, wystarczające przy skali MVP–kilkuset placówek. Każde zapytanie musi filtrować po `organization_id` (najlepiej wymuszone przez middleware + Row-Level Security w PostgreSQL).

---

## 4. Obiekty bazy danych

### 4.1 `organizations` (placówki)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| name | text | Nazwa placówki |
| type | enum | `przedszkole`, `zlobek`, `kolko` |
| subdomain | text unique | np. `sloneczko.twojaapka.pl` |
| address | text | |
| nip | text | |
| logo_url | text | |
| primary_color | text | branding per placówka |
| subscription_status | enum | `trial`, `active`, `past_due`, `cancelled` |
| subscription_plan | text | |
| created_at | timestamptz | |

### 4.2 `users` (konta — rodzice, personel, admini)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| organization_id | uuid FK → organizations | |
| email | text unique | |
| password_hash | text | |
| phone | text | |
| role | enum | `parent`, `teacher`, `admin`, `owner` |
| first_name / last_name | text | |
| push_token | text nullable | ostatni znany token FCM (lub osobna tabela — patrz 4.14) |
| is_active | bool | |
| created_at | timestamptz | |

### 4.3 `groups` (grupy przedszkolne / oddziały)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| organization_id | uuid FK | |
| name | text | np. "Biedronki" |
| teacher_id | uuid FK → users nullable | |
| age_range | text | |

### 4.4 `children`
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| organization_id | uuid FK | |
| group_id | uuid FK → groups | |
| first_name / last_name | text | |
| birth_date | date | |
| notes | text nullable | ograniczone do informacji logistycznych — bez dziennika rozwojowego |
| created_at | timestamptz | |

### 4.5 `parent_child` (relacja rodzic–dziecko, wiele-do-wielu)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| parent_id | uuid FK → users | |
| child_id | uuid FK → children | |
| relation | enum | `matka`, `ojciec`, `opiekun` |
| can_pickup | bool | czy może odbierać dziecko |

### 4.6 `announcements` (ogłoszenia)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| organization_id | uuid FK | |
| group_id | uuid FK nullable | null = ogłoszenie do całej placówki |
| author_id | uuid FK → users | |
| title | text | |
| content | text | |
| attachment_url | text nullable | |
| published_at | timestamptz | |

### 4.7 `conversations` + `messages` (czat)
**conversations**
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| organization_id | uuid FK | |
| parent_id | uuid FK → users | |
| teacher_id | uuid FK → users nullable | |
| child_id | uuid FK nullable | kontekst rozmowy |
| created_at | timestamptz | |

**messages**
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| conversation_id | uuid FK | |
| sender_id | uuid FK → users | |
| content | text | |
| attachment_url | text nullable | |
| read_at | timestamptz nullable | |
| created_at | timestamptz | |

### 4.8 `absences` (nieobecności)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| child_id | uuid FK | |
| reported_by | uuid FK → users | |
| date_from | date | |
| date_to | date | |
| reason | text nullable | |
| created_at | timestamptz | |

### 4.9 `events` (wydarzenia)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| organization_id | uuid FK | |
| group_id | uuid FK nullable | |
| title | text | |
| description | text | |
| requires_consent | bool | np. zgoda na wycieczkę |
| requires_item | text nullable | np. "przynieś kasztany" |
| event_date | timestamptz | |
| registration_deadline | timestamptz nullable | |
| created_by | uuid FK → users | |

### 4.10 `event_registrations`
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| event_id | uuid FK | |
| child_id | uuid FK | |
| confirmed_by | uuid FK → users | |
| status | enum | `confirmed`, `declined`, `pending` |
| consent_given | bool | |
| created_at | timestamptz | |

### 4.11 `payments`
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| organization_id | uuid FK | |
| child_id | uuid FK | |
| payer_id | uuid FK → users | |
| type | enum | `posilki`, `zajecia`, `wydarzenie`, `inne` |
| amount_cents | integer | kwoty zawsze w groszach |
| currency | text | domyślnie `PLN` |
| status | enum | `pending`, `paid`, `failed`, `refunded` |
| provider | enum | `tpay`, `przelewy24` |
| provider_transaction_id | text nullable | |
| due_date | date nullable | |
| paid_at | timestamptz nullable | |
| created_at | timestamptz | |

### 4.12 `daily_schedule` (plan dnia grupy)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| group_id | uuid FK | |
| date | date | |
| time_slot | time | |
| activity | text | np. "śniadanie", "zajęcia plastyczne" |
| created_by | uuid FK → users | |

### 4.13 `notifications` (log wysłanych powiadomień)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK | |
| type | enum | `push`, `email`, `sms` |
| title | text | |
| body | text | |
| related_entity | text nullable | np. `announcement:uuid` |
| sent_at | timestamptz | |
| read_at | timestamptz nullable | |

### 4.14 `push_tokens`
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK | |
| token | text | token FCM |
| platform | enum | `ios`, `android`, `web` |
| created_at | timestamptz | |

### 4.15 `audit_log` (rekomendowane od początku, RODO)
| Kolumna | Typ | Opis |
|---|---|---|
| id | uuid PK | |
| organization_id | uuid FK | |
| actor_id | uuid FK → users | |
| action | text | np. `child.updated`, `payment.refunded` |
| entity_type / entity_id | text / uuid | |
| metadata | jsonb | |
| created_at | timestamptz | |

---

## 5. Szkielet endpointów API (Go, REST)

```
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/register

GET    /api/v1/children
GET    /api/v1/children/:id

POST   /api/v1/absences
GET    /api/v1/absences?child_id=

GET    /api/v1/announcements
POST   /api/v1/announcements          (teacher/admin)

GET    /api/v1/conversations
POST   /api/v1/conversations/:id/messages
GET    /api/v1/conversations/:id/messages   (paginacja)

GET    /api/v1/events
POST   /api/v1/events/:id/register

POST   /api/v1/payments/initiate
POST   /api/v1/payments/webhook/tpay        (publiczny, weryfikacja podpisu)
GET    /api/v1/payments?child_id=

GET    /api/v1/groups/:id/schedule?date=
```

Real-time czatu: WebSocket (`/ws`) lub Server-Sent Events — polecam WebSocket, bo już masz w tym doświadczenie z projektu klubu tenisowego.

---

## 6. Kolejność wdrażania (sugerowana)

1. `organizations`, `users`, `children`, `groups`, `parent_child` + auth
2. `announcements` + push (FCM)
3. `absences`
4. `conversations` + `messages` (czat)
5. `events` + `event_registrations`
6. `payments` (Tpay/Przelewy24) + `notifications`
7. `daily_schedule`
8. `audit_log` (dociągnąć w tle, ale nie odkładać na koniec)
