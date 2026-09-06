# Podsumowanie projektu: Komunikator dla przedszkoli, żłobków i kółek pozalekcyjnych

## 1. Na czym polega projekt

**Co to jest:** aplikacja typu multi-tenant (jedna baza kodu, wielu klientów z własnym brandingiem) skierowana do:
- przedszkoli i żłobków (małych, głównie prywatnych, bez obecnego systemu),
- kółek pozalekcyjnych i zajęć dodatkowych nieprowadzonych przez szkołę (szkoły tańca, sportu, języków, korepetycje, akademie dziecięce).

**Czego produkt NIE robi (świadomie):** nie jest dziennikiem elektronicznym w rozumieniu przepisów MEN (dokumentacja pedagogiczna, archiwizacja podpisem z profilu zaufanego, zgoda organu prowadzącego). Placówka zachowuje swój obecny system dziennika (Librus, Vulcan/UONET+, mobiDziennik lub papier) bez zmian — produkt działa obok niego jako dodatkowy kanał.

**Kluczowe funkcje MVP:**
- komunikacja z rodzicami (czat, ogłoszenia, powiadomienia push),
- zgłaszanie nieobecności przez rodzica jednym kliknięciem,
- zapisy na wydarzenia z potwierdzeniem uczestnictwa (np. zgoda na wycieczkę, "przynieś kasztany"),
- płatności online za posiłki/zajęcia (BLIK, Przelewy24 lub Tpay),
- informacyjny plan/harmonogram dnia grupy (bez indywidualnych notatek rozwojowych dziecka — to pozostaje poza zakresem, żeby nie wejść w definicję dziennika).

**Model techniczny:** Go (backend/API) + SvelteKit (frontend, PWA na start — własna ikona i branding per subdomena bez czekania na App Store) + docelowo natywna apka KMP jako jedna, wspólna aplikacja multi-tenant z dynamicznym logo/nazwą po wyborze placówki.

**Model sprzedaży:** osobiste wizyty, start na Śląsku z rozszerzeniem na całą Polskę, pilotaż 2-3 miesiące za darmo dla pierwszych, referencyjnych klientów (masz już dwóch znanych dyrektorów placówek), później 30 dni za darmo standardowo.

---

## 2. Ocena projektu według kategorii

| Kategoria | Ocena | Komentarz |
|---|---|---|
| **Wielkość rynku** | Dobra | 21,9 tys. placówek przedszkolnych w Polsce + żłobki + rynek kółek pozalekcyjnych (osobny, duży segment) |
| **Konkurencja — przedszkola/żłobki** | Trudna | LiveKid (finansowanie VC, 11 krajów, 5000+ placówek) + 6-7 mniejszych graczy (Inso, eLizak, KidsTime, Kidplace, kidsup, Pan Licznik) |
| **Konkurencja — kółka pozalekcyjne** | Umiarkowana | Kilku graczy (Klasera, Classivo, ClassPilot, Zyko, LuppoSystem), ale bootstrapowani, bez kapitału VC — Twojej wagi |
| **Bariera regulacyjna** | Niska (przy obranym zakresie) | Świadomie pomijasz dziennik pedagogiczny (MEN/profil zaufany) — to usuwa największą barierę wejścia |
| **Dopasowanie do Twoich umiejętności** | Bardzo dobre | Go + SvelteKit + PWA/KMP w pełni wystarczające na cały zakres MVP |
| **Łatwość sprzedaży osobistej** | Dobra | Masz już 2 realne kontakty do pilotażu; model "taniej niż LiveKid + osobisty kontakt" jest wiarygodny dla mikro-placówek |
| **Ryzyko dwustronnej adopcji** | Niskie | Rodzic używa apki wielokrotnie w tygodniu (nieobecności, płatności, ogłoszenia) — częsty kontakt ułatwia przyzwyczajenie |
| **Ryzyko finansowe przy 20 tys. zł kapitału** | Akceptowalne | Głównie koszt Twojego czasu, nie kapitału — niskie koszty stałe jako JDG |
| **Realność celu "główne źródło dochodu za rok"** | Zbyt optymistyczne przy obecnych założeniach | Wymaga wyższej ceny, szybszego tempa sprzedaży lub dłuższego horyzontu (patrz projekcja z poprzedniej wiadomości) |

**Ogólna ocena: projekt ma realną szansę, pod warunkiem trzymania wąskiego zakresu (bez dziennika) i priorytetu na kółka pozalekcyjne jako łatwiejszy pierwszy segment przychodowy, z przedszkolami/żłobkami jako równoległym, dłuższym torem sprzedażowym.**

---

## 3. Proponowany cennik

| Segment | Cena | Uwagi |
|---|---|---|
| Przedszkole/żłobek (do ok. 40-50 dzieci) | **79-99 zł/mies.** | Ryczałt, nie per dziecko na start |
| Kółko pozalekcyjne / szkoła zajęć dodatkowych | **79-129 zł/mies.** | Zależnie od liczby uczestników |
| Pierwsi 2-3 klienci referencyjni (pilotaż) | **0 zł przez 2-3 miesiące** | W zamian za feedback i zgodę na referencję |
| Każdy kolejny nowy klient | **0 zł przez 30 dni**, potem cena standardowa | Krótszy okres próbny niż dla pilotażu |
| Płatność roczna z góry (opcjonalnie, później) | **rabat 10-15%** | Wprowadzić dopiero po ustabilizowaniu produktu |

---

## 4. Szacowany czas budowy i etapy

Założenie: do 6h dziennie, realistycznie nie każdego dnia (masz inne źródło dochodu) — przyjęto ok. 25-30h/tydzień efektywnej pracy.

| Etap | Zakres | Szacowany czas |
|---|---|---|
| 1. Walidacja | Rozmowy z 2 znanymi dyrektorami, doprecyzowanie zakresu funkcji | 3-5 dni |
| 2. Projekt architektury | Model danych multi-tenant, struktura API, plan subdomen/brandingu | 1 tydzień |
| 3. Backend (Go) | Autoryzacja, zarządzanie tenantami, API: komunikacja, nieobecności, wydarzenia | 3-4 tygodnie |
| 4. Frontend PWA (SvelteKit) | Panel rodzica + panel administracyjny placówki, manifest.json per subdomena | 3-4 tygodnie |
| 5. Integracja płatności | BLIK/Przelewy24/Tpay do płatności za posiłki/zajęcia | 1 tydzień |
| 6. Testy wewnętrzne | Testowanie całego przepływu przed pilotażem | 1 tydzień |
| 7. Wdrożenie pilotażowe | Uruchomienie u 2 dyrektorów, zbieranie feedbacku | 2-3 tygodnie (równolegle z bieżącym użytkowaniem) |
| 8. Iteracja po pilotażu | Poprawki na podstawie realnych uwag przed dalszą sprzedażą | 2-3 tygodnie |

**Łączny czas do wersji gotowej na szerszą sprzedaż: ok. 3-4 miesiące** przy założonym tempie pracy.

Etap dodatkowy (later, po walidacji modelu): natywna aplikacja mobilna multi-tenant w KMP z dynamicznym logo/nazwą — osobny projekt po potwierdzeniu, że model PWA się sprawdza, szacunkowo kolejne 4-6 tygodni pracy.

---

*Dokument przygotowany na podstawie analizy rynku i rozmowy strategicznej. Liczby dotyczące tempa sprzedaży i przychodów są szacunkami ilustracyjnymi — wymagają weryfikacji w realnym pilotażu.*
