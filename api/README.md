# Scrabble API — Backend (Go + Echo)

API REST pour un jeu de Scrabble en ligne. Gestion des comptes, parties, coups, scores, suggestions d’utilisateurs, système de reports, notifications push, et migrations PostgreSQL.

## Fonctionnalités

* Authentification JWT (login/register) + rôle `admin` pour certaines routes.
* Création/gestion de parties multijoueurs (plateau 15×15, pioche, racks, score, historique des coups).
* Validation des mots via dictionnaire FR embarqué (accents gérés) et règles de placement (centre, continuité, connexion).
* Calcul du score avec cases spéciales (DL, TL, DW, TW, ★ centre).
* Fin de partie quand un joueur vide son rack avec sac vide *ou* après `2 × nombre de joueurs` passes consécutives.
* Système de signalements (reports) avec statut/priorité/type + endpoints admin.
* Suggestions d’utilisateurs (auto‑complétion par préfixe).
* Notifications Web Push (VAPID) : abonnement côté API, envoi de notifications.
* Migrations de schéma avec Goose.

---

## Stack & Architecture

* **Langage** : Go
* **Web framework** : Echo
* **DB** : PostgreSQL
* **Migrations** : Goose (SQL + helpers Go)
* **Auth** : JWT (HMAC, secret env)
* **Sécurité** : middlewares RequireAuth/RequireAdmin
* **Logs** : zap
* **Push** : webpush‑go (VAPID)

### Organisation du code

```
api/
  controller/   # Handlers HTTP (auth, game, report, users, notifications)
  routes/       # Déclaration des routes et middlewares
  services/     # Logique métier et accès DB
  models/       # DTO request/response + modèles DB
  middleware/   # JWT + contrôle d’accès admin
  utils/        # Helpers (JWT, notif, lettres, context)
  word/         # Dictionnaire FR embarqué + grille spéciale
migrations/     # Fichiers SQL + runners Go (up/down)
```

---

## Données & Migrations

Schéma principal :

* `users` : comptes (username unique, password hashé, rôle).
* `games` : parties (UUID, créateur, statut, tour courant, plateau JSON, sac de lettres, `pass_count`, `winner_username`, `ended_at`).
* `game_players` : participants (rack, position, score).
* `game_moves` : historique des coups (JSONB, horodatage).
* `reports` : signalements (titre, contenu, statut/priorité/type, timestamps).
* `push_subscriptions` : abonnement Web Push par utilisateur.

### Lancer les migrations

1. Copier la config d’exemple puis renseigner la connexion :

```bash
cp api/.env.example api/.env
# Éditer POSTGRES_URL, JWT_SECRET, VAPID_*
```

2. Démarrer Postgres :

```bash
make db
```

3. Exécuter les migrations *up* :

```bash
make migrate-up
```

> Pour revenir en arrière : `make migrate-down`.

---

## Configuration

Variables d’environnement (fichier `api/.env`) :

* **Base de données** : `POSTGRES_URL` (ou HOST/PORT/USER/PASSWORD/DB)
* **JWT** : `JWT_SECRET`
* **Web Push** : `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY`
* **Logs** : `LOGS_PASSWORD` (optionnel)

> Un exemple est fourni dans `api/.env.example`. Pensez à ne **pas** commiter votre `.env`.

---

## Lancer l’API en dev

### Avec live‑reload

```bash
make air
# par défaut, l’API écoute sur :8888
```

### Sans live‑reload

```bash
cd api && go run .
```

## Conteneur Docker (optionnel)

Un `Dockerfile` est fourni dans `api/`. Vous pouvez builder et exécuter l’API, ou décommenter le service `api` dans `docker-compose.yml` pour démarrer DB + API + front.

---

## Authentification

* JWT signé côté serveur (algorithme HMAC) via `JWT_SECRET`.
* Ajouter le header `Authorization: Bearer <token>` sur toutes les routes protégées.
* Rôle `admin` requis pour les routes d’administration (reports).

### Flux de base

1. **Register** → crée un utilisateur.
2. **Login** → renvoie un token JWT à stocker côté client.
3. Utiliser ce token pour appeler les routes `/game`, `/users`, `/report`, `/notifications`.

---

## Endpoints

> Les données de réponse sont simplifiées ci‑dessous. Voir les sections *Models* pour les structures complètes.

### Auth

* `POST /auth/register` `{ username, password }` → crée un compte.
* `POST /auth/login` `{ username, password }` → `{ token }`.
* `POST /auth/change-password` *(admin + auth)* `{ username, new_password }` → 200.
* `GET  /auth/connect-as?user=<username>` → `{ token }` (outil pratique d’admin/dev).

### Jeux

* `POST /game` *(auth)*

  * body : `{ name: string, players: string[] }` (usernames invités)
  * crée la partie, attribue les racks, set `current_turn` au créateur.
* `GET /game` *(auth)* → liste des parties de l’utilisateur (avec dernier coup, tour courant, propriétaire, gagnant si terminé).
* `GET /game/:id` *(auth)* → détails complets : plateau, votre rack, joueurs, historique, statut, lettres restantes.
* `PUT /game/:id/rename` *(créateur)* `{ new_name }` → renomme la partie.
* `DELETE /game/:id` *(créateur)* → supprime partie + joueurs + coups.
* `POST /game/:id/play` *(tour courant)*

  * body : `{ letters: [{x,y,char,blank?}, ...] }` (`blank` est optionnel; si omis, l’API déduira l’usage d’un joker `?` si nécessaire depuis votre rack)
  * contraintes : 1 seule ligne/colonne, 1er coup couvre le centre, connexion aux lettres existantes, lettres doivent être dans le rack, max 7 posées.
  * effets : met à jour le plateau, calcule/ajoute le score, recharge le rack, sauvegarde le coup, remet `pass_count=0`, passe au joueur suivant.
  * fin de partie si joueur vide son rack **et** sac vide → `winner_username` + `ended_at`.
* `POST /game/:id/pass` *(tour courant)*

  * enregistre un « pass », passe au joueur suivant, incrémente `pass_count`.
  * fin de partie si `pass_count >= 2 × nb_joueurs`.
* `GET /game/:id/new_rack` *(tour courant)*

  * échange intégral du rack : tire 7 nouvelles lettres (si sac non vide), remet l’ancien rack dans le sac, puis passe au joueur suivant.
* `POST /game/:id/simulate_score` *(auth)*

  * body : `{ letters: [{x,y,char}, ...] }`
  * renvoie `{ score }` sans modifier l’état.

### Reports (signalements)

* `POST /report` *(auth)* `{ title, content }` → crée un report.
* `GET  /report/me` *(auth)* → mes reports.
* `GET  /report/:id` *(auth)* → détails d’un report.
* `GET    /report` *(admin)* → tous les reports (ordre chronologique inverse).
* `PATCH  /report/:id` *(admin)* → modifie `title`, `content`, `status`.
* `PUT    /report/:id/resolve|reject|progress` *(admin)* → change le statut.
* `DELETE /report/:id` *(admin)* → supprime un report.

### Utilisateurs

* `GET /users/suggest?q=<prefix>` *(auth)* → top 10 usernames correspondant au préfixe.

### Notifications

* `POST /notifications/push-subscribe` *(auth)* → enregistre/maj l’abonnement push (VAPID).
* `GET  /notifications/test` → envoie une notif de test à un utilisateur (endpoint de debug).

---

## Modèles (DTO)

### Requests

* **Auth** : `RegisterRequest { username, password }`, `LoginRequest { username, password }`, `ChangePasswordRequest { username, new_password }`.
* **Game** : `CreateGameRequest { name, players }`, `RenameGameRequest { new_name }`, `PlacedLetter { x, y, char, blank? }`, `PlayMoveRequest { letters: PlacedLetter[] }`.
* **Report** : `CreateReportRequest { title, content }`, `UpdateReportRequest { title?, content?, status? }`.

### Responses

* **Auth** : `AuthResponse { token }`.
* **Game** : `GameInfo` (plateau + rack + joueurs + coups + statut), `GameSummary` (liste), `MoveInfo`, `PlayerInfo`.
* **Users** : `SuggestUsersResponse { id, username }`.
* **Report** : `Report` (inclut `username` au lieu de l’ID).

---

## Règles du jeu implémentées

* **Plateau** : 15×15, cases spéciales : `DL`, `TL`, `DW`, `TW`, `★` au centre.
* **Dictionnaire** : fr.txt embarqué, mots normalisés (majuscules, accents supprimés) pour la validation.
* **Placement** : premier mot couvre le centre ; ensuite, continuité et connexion obligatoires.
* **Score** : somme des lettres (valeurs FR) avec multiplicateurs de **lettre** et **mot** selon les cases traversées. Bonus de 7 lettres (bingo) si applicable. Les deux jokers valent 0 point et n'obtiennent aucun multiplicateur de lettre.
* **Fin de partie** :

  * soit un joueur pose son dernier jeton **et** le sac est vide ;
  * soit `pass_count >= 2 × nb_joueurs`.
* **Notifications** : après chaque coup ou fin de partie, l’adversaire/le gagnant reçoit une notif (si abonné).

---

## Exemples (cURL)

> Remplacez `$API` par l’URL de l’API (ex. `http://localhost:8888`).

### Register & Login

```bash
curl -X POST "$API/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"pwd"}'

curl -X POST "$API/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"pwd"}'
# -> { "token": "…" }
```

### Créer une partie

```bash
TOKEN=…
curl -X POST "$API/game" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Vendredi soir","players":["bob","carol"]}'
```

### Jouer un coup

```bash
curl -X POST "$API/game/$GAME_ID/play" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"letters":[{"x":7,"y":7,"char":"C"},{"x":8,"y":7,"char":"A","blank":true}]}'
```

### Passer son tour

```bash
curl -X POST "$API/game/$GAME_ID/pass" -H "Authorization: Bearer $TOKEN"
```

### Simuler un score (sans modifier l’état)

```bash
curl -X POST "$API/game/$GAME_ID/simulate_score" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"letters":[{"x":7,"y":7,"char":"C"},{"x":8,"y":7,"char":"A","blank":true}]}'
# -> { "score": 24 }
```

### Reports

```bash
curl -X POST "$API/report" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Bug plateau","content":"Case bloquée"}'
```

---

## Tests

Exemples de tests unitaires : validation des mots (`api/word/word_test.go`).

```bash
cd api && go test ./...
```

---

## Cibles Make utiles

* `make db` : démarre Postgres via Docker Compose.
* `make migrate-up` / `make migrate-down` : applique/revert les migrations sur `POSTGRES_URL`.
* `make air` : live‑reload API.
* `make front` : démarre le frontend (si présent).

---
