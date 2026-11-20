# Baza danych Rodzic Chat — obiekty i ich rola

Ten plik opisuje wszystkie tabele (obiekty) w bazie danych backendu, po co istnieją i jak się ze sobą łączą. Backend używa GORM (PostgreSQL) z auto-migracją — nowe pola/tabele pojawiają się w bazie automatycznie przy starcie serwera, bez ręcznych plików migracyjnych.

## Konwencje ogólne

- **Id** — wszędzie losowy string (UUID), nigdy liczba porządkowa.
- **Kwoty pieniężne** — zawsze w groszach (`int64`), np. `150000` = 1500,00 zł. Nigdy `float`, żeby uniknąć błędów zaokrągleń.
- **Enumy** — przechowywane jako `int` (kolejny numer w stałej liście). **Nowe wartości enuma trzeba zawsze dopisywać na końcu listy, nigdy wstawiać w środku** — inaczej po cichu zmieni się znaczenie już zapisanych w bazie wierszy.
- **Miękkie usuwanie (soft-delete)** — tylko `Account` ma `DeletedAt` (konwencja GORM): usunięte konto znika z normalnych zapytań, ale zostaje w bazie. Wszystkie inne tabele usuwają się na twardo.
- **Wielotenancy** — prawie każda tabela ma `OrganizationId` — dane jednej placówki nigdy nie mieszają się z danymi innej.
- **CreateTime / UpdateTime** — na każdej tabeli, ustawiane ręcznie w kodzie (nie przez triggery bazy danych).

---

## 1. Konta, placówki, role

### `Organization` (`organizations`)
Jedna placówka (przedszkole/żłobek/szkoła/kółko) — główny "tenant" całego systemu.
- `Type` — Kindergarten / Nursery / School / Extracurricular / Independent.
- `Subdomain` — unikalna, np. `mojeprzedszkole.rodzicchat.pl`.
- `SubscriptionStatus` — Trial / Active / PastDue / Cancelled (obecnie **zapisywane, ale nigdzie jeszcze nie egzekwowane** — patrz sekcja "Do dokończenia" niżej).
- `ChildrenLimit` — wybrany przy rejestracji próg cenowy (1–20, 21–30 … 401–500 dzieci); `0` = placówka sprzed wprowadzenia tego pola, traktowana jako bez limitu.
- `PaymentReference` — unikalny, wygenerowany kod (np. `RC-7F3K9Q`) używany jako tytuł przelewu za abonament platformy.
- `SubscriptionFreeOverride` — zwalnia placówkę z opłaty; ustawia to wyłącznie operator (Ty), nigdy dyrektor.

### `Account` (`accounts`)
Każdy użytkownik logujący się do systemu — rodzic, nauczyciel, dyrektor, współdyrektor.
- `AccountType` — Parent / Teacher / Director / CoDirector.
- `Status` — Enabled / Disabled / Banned.
- `Login` — unikalny (zwykle e-mail).
- Ma `DeletedAt` (jedyna tabela z miękkim usuwaniem).

### `Credential` (`credentials`)
Osobno od `Account` — hasło (zahaszowane bcrypt) powiązane z `UserId`. Rozdzielone od `Account` prawdopodobnie po to, by dane logowania trzymać osobno od danych profilowych.
> ⚠️ Uwaga techniczna (zweryfikowane bezpośrednio w bazie): ta tabela **faktycznie nie ma żadnego klucza głównego** — ani jawnego, ani automatycznie dołożonego przez GORM. Działa to dziś tylko dlatego, że kod zawsze filtruje/aktualizuje po `user_id`, ale nic w samej bazie nie wymusza unikalności tego pola — teoretycznie dało(by) się wstawić dwa wiersze `Credential` dla tego samego konta. Warto to naprawić przy najbliższej okazji dotykania tej tabeli (dodać `gorm:"primaryKey"` na `UserId` i przemigrować).

### `Session` (`sessions`)
Jedno zalogowane urządzenie/sesja. Przechowuje zahaszowany refresh token (`RefreshTokenHash`), dzięki czemu `/refresh-token` może go zweryfikować i podmienić bez trzymania go w postaci jawnej. `RevokedAt` pozwala wylogować pojedyncze urządzenie albo unieważnić wszystkie sesje (np. przy resecie hasła).

### `AccountOrganization` (`accountorganizations`)
Członkostwo **rodzica** w placówce — rodzic może mieć dzieci w kilku placówkach naraz i przełączać się między nimi. Dyrektor/nauczyciel zawsze należy tylko do jednej (przez `Account.OrganizationId`).

### `PasswordReset` (`passwordresets`)
Jednorazowy, krótko ważny link do resetu hasła. W bazie trzymany jest tylko hash tokenu (SHA-256) — sam token w treści maila nie jest odtwarzalny z bazy. `UsedAt` blokuje ponowne użycie tego samego linku.

---

## 2. Dzieci, grupy, relacje rodzinne

### `Child` (`children`)
Dziecko zapisane do placówki. Należy do jednej `Group` (opcjonalnie — może nie mieć jeszcze przypisanej grupy).

### `Group` (`groups`)
Grupa/klasa w placówce. `TeacherId` to "główny" nauczyciel; `RequireChildBirthDateVerification` włącza dodatkową weryfikację przy samodzielnej rejestracji rodzica przez link (patrz `GuardianClaim` niżej) — zabezpieczenie przed podszywaniem się pod rodzica kolegi z grupy.

### `GroupTeacher` (`groupteachers`)
Dodatkowi nauczyciele przypisani do grupy — uzupełnia `Group.TeacherId`. Pełna lista nauczycieli grupy = `Group.TeacherId` + wszystkie wiersze `GroupTeacher` dla tej grupy.

### `ParentChild` (`parentchild`)
Łączy rodzica (`ParentId`) z dzieckiem (`ChildId`). `Relation` (Matka/Ojciec/Opiekun) i `CanPickup` (czy ta osoba może odebrać dziecko) — to na tej tabeli opiera się cała logika "kto widzi dane tego dziecka".

### `GuardianClaim` (`guardianclaims`)
Kolejka do zatwierdzenia przez dyrektora, gdy drugi rodzic zgłasza się do dziecka przez link rejestracyjny, ale podana data urodzenia dziecka **nie zgadza się** z tym, co jest w systemie. Zamiast automatycznie łączyć konto z dzieckiem (co pozwoliłoby komuś obcemu, kto zna tylko imię i nazwisko, podszyć się pod rodzica), tworzy się ten wiersz do ręcznej weryfikacji. Zawiera migawkę danych z momentu zgłoszenia (`ClaimantName`, `ClaimantEmail`, `ClaimantBirthDate`, `ChildBirthDateOnFile`), żeby dyrektor mógł je porównać nawet jeśli dane w międzyczasie się zmienią.

---

## 3. Czat, ogłoszenia, ankiety

### `Channel` (`channels`)
Kanał czatu. `Type` decyduje o tym, kto może pisać:
- **Announcement** (ogłoszenia) — pisze wyłącznie kadra (dyrektor/współdyrektor/nauczyciel).
- **Group** — kadra może pisać zawsze; zasady dla rodziców zależą od konfiguracji.
- **Direct** — czat 1:1.
- **Parents** ("rodzice między sobą") — inaczej niż w kanale grupowym, kadra **nie** ma tu automatycznego prawa pisania bez warunków — to czat rodziców nawzajem.

### `Message` (`messages`)
Pojedyncza wiadomość. `Type` = Text / Poll / System. Ankieta (`Poll`) to technicznie też wiadomość — `Type=Poll` — dlatego dziedziczy te same zasady uprawnień do pisania co zwykła wiadomość w danym typie kanału.

### `MessageRead` (`message_reads`)
**Potwierdzenie odczytu.** Jeden wiersz na parę (wiadomość, konto) — unikalność wymuszona indeksem `idx_message_read`. Ponowne oznaczenie jako przeczytane tylko aktualizuje `ReadAt`, nie duplikuje wiersza.

### `MessageReaction` (`message_reactions`)
Lekka reakcja emoji na wiadomość (żeby rodzic nie musiał pisać "OK" pod każdym ogłoszeniem). Jeden wiersz na (wiadomość, konto) — nowa reakcja nadpisuje poprzednią.

### `Poll`, `PollOption`, `PollVote` (`polls`, `poll_options`, `poll_votes`)
Ankieta powiązana z konkretną wiadomością (`Poll.MessageId`). `AllowMultiple` pozwala zaznaczyć więcej niż jedną opcję. `ClosesAt` — opcjonalny termin zamknięcia głosowania. Każdy głos to osobny wiersz w `PollVote` (poll_id + option_id + account_id) — więc łatwo policzyć wyniki i sprawdzić, kto jak zagłosował.

---

## 4. Wydarzenia, nieobecności, obiady

### `Event` (`events`)
Wydarzenie (wycieczka, uroczystość itp.). `RequiresConsent` + `RequiresItem` pozwalają zaznaczyć, że wydarzenie wymaga zgody rodzica i/lub przyniesienia czegoś. `RegistrationDeadline` — opcjonalny termin zapisów.

### `EventRegistration` (`event_registrations`)
Zapis/potwierdzenie udziału dziecka w wydarzeniu. `Status` = Confirmed / Declined / Pending, `ConsentGiven` — czy rodzic potwierdził zgodę. To jest odpowiednik "kto potwierdził udział" dla wydarzeń (patrz sekcja niżej o odczytach/interakcjach).

### `Absence` (`absences`)
Zgłoszenie nieobecności dziecka (`DateFrom`–`DateTo`, opcjonalny `Reason`) zgłoszone przez rodzica lub kadrę (`ReportedBy`).

### `MealCancellation` (`meal_cancellations`)
To samo co `Absence`, ale dla rezygnacji z posiłków — osobna tabela, bo ma inne znaczenie biznesowe (rozliczenia z kuchnią), mimo identycznego kształtu.

---

## 5. Płatności (rodzic → placówka, przez Stripe)

### `Payment` (`payments`)
Jednorazowa opłata przypisana przez dyrektora jednemu rodzicowi za jedno dziecko (czesne, obiady, zajęcia, wycieczka, inne — `Category`). Przypisanie do wielu rodziców naraz tworzy kilka wierszy powiązanych wspólnym `BatchId`.
- `BaseAmountGrosze` — ustalane raz, to dokładnie tyle, ile ma dostać placówka.
- `PlatformFeeGrosze`, `StripeFeeEstimateGrosze`, `TotalAmountGrosze` — liczone dopiero w momencie, gdy rodzic zaczyna płatność online (przy przelewie bankowym zostają zerowe, bo Stripe w ogóle ich nie widzi).
- `PaymentMethod` — Online (BLIK przez Stripe) / BankTransfer (ręcznie zaksięgowane przez kadrę).
- `Status` — Unpaid / Paid / Reconciled. Paid ustawia się automatycznie po udanej płatności BLIK (webhook Stripe); Reconciled ustawia ręcznie kadra dla przelewu bankowego, który Stripe nigdy nie widzi.

### `OrganizationStripeAccount` (`organization_stripe_accounts`)
Łączy placówkę z jej kontem Stripe Connect (Express) — tam trafiają wpłaty rodziców pomniejszone o prowizje. `ChargesEnabled`/`PayoutsEnabled`/`DetailsSubmitted` to lustrzane odbicie stanu konta w Stripe, aktualizowane webhookiem `account.updated`.

### `PlatformSettings` (`platform_settings`)
Jeden, globalny wiersz (o stałym `Id = "platform-settings"`) z ustawieniami finansowymi: Twoja marża od płatności online (`PlatformFeeGrosze`) i założona prowizja Stripe używana do wyliczenia kwoty pobieranej od rodzica z góry (`StripeAssumedPercentBps`, `StripeAssumedFixedGrosze`).

---

## 6. Zaproszenia i samodzielna rejestracja

### `RegistrationLink` (`registration_links`)
Link, którym dyrektor zaprasza rodziców jednej grupy do samodzielnej rejestracji. Limit użyć = `ChildrenCount * 2` (maks. 2 rodziców na dziecko), wygasa po 30 dniach. Brak osobnego pola statusu — "wygasły"/"wyczerpany" liczy się w locie z `ExpiresAt`/`UsedCount`. Odwołanie linku = twarde usunięcie wiersza.

### `TeacherInvite` (`teacher_invites`)
To samo, ale dla zaproszenia nauczyciela (albo współdyrektora — `AccountType` = `teacher` lub `co_director`; tylko Dyrektor, nie CoDirector, może zapraszać współdyrektora).

---

## 7. Pliki i uprawnienia

### `Media` (`media`)
Plik/zdjęcie wgrane do systemu (galeria, załączniki do wydarzeń typu zgoda na wycieczkę). `ObjectKey` to losowa ścieżka w magazynie (Cloudflare R2) — nie nazwa pliku; `OriginalFilename` to co faktycznie wysłała przeglądarka, żeby dało się sensownie pokazać/pobrać plik. `Category` to dowolny tekst-podpowiedź do filtrowania w galerii mobilnej (np. „schedule”, „meal_plan”) — backend go nie waliduje.

### `OrganizationFeature` (`organization_features`)
Włącznik/wyłącznik funkcji (`Feature`: Chat, Announcements, Absences, Events, Gallery, MealCancellations, Payments) per placówka. Jeden wiersz na (organizacja, funkcja) — to na tej tabeli opiera się, czy dana placówka w ogóle widzi np. moduł Płatności.

---

## Enumy — ściągawka

| Enum | Wartości |
|---|---|
| `AccountType` | Parent, Teacher, Director, CoDirector |
| `Account.Status` | Enabled, Disabled, Banned |
| `OrganizationType` | Kindergarten, Nursery, School, Extracurricular, Independent |
| `Organization.SubscriptionStatus` | Trial, Active, PastDue, Cancelled |
| `ChannelType` | Announcement, Group, Direct, Parents |
| `MessageType` | Text, Poll, System |
| `EventType` | General, Trip |
| `EventRegistration.Status` | Confirmed, Declined, Pending |
| `GuardianClaimStatus` | Pending, Approved, Rejected |
| `PaymentCategory` | Tuition (czesne), Meals (obiady), Extracurricular (zajęcia), Trip (wycieczka), Other (inne) |
| `Payment.Status` | Unpaid, Paid, Reconciled |
| `PaymentMethod` | Unset, Online, BankTransfer |
| `RelationType` (parentchild) | Mother, Father, Guardian |
| `Feature` | Chat, Announcements, Absences, Events, Gallery, MealCancellations, Payments |

---

## Do dokończenia / znane luki

- **`Organization.SubscriptionStatus` nie jest jeszcze nigdzie egzekwowany** — pole istnieje i domyślnie ustawia się na `Trial` przy rejestracji, ale nic w kodzie nie sprawdza, czy status wygasł, ani nie blokuje dostępu placówce, która nie zapłaciła. To osobna, jeszcze nierozwiązana decyzja produktowa (długość triala, okres karencji, co dokładnie się blokuje).
- **`Credential` bez żadnego klucza głównego** (zweryfikowane w bazie: `\d credentials`) — patrz sekcja 1. Nic dziś nie blokuje dwóch wierszy dla jednego `user_id`.
