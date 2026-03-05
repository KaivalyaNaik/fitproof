
# Fitness Challenge Platform – System Design

## 1. Problem Statement

A group of participants is running a fitness challenge where members must complete daily fitness goals and submit proof of completion. Currently, progress is tracked manually using shared spreadsheets, which is time-consuming, error-prone, and difficult to manage as the group grows.

The goal of this system is to provide an automated and scalable platform that allows users to participate in fitness challenges, submit daily activity metrics, automatically calculate points and penalties, and maintain a real-time leaderboard to motivate participants.

---

# 2. Goals

1. Authenticated users should be able to create new fitness challenges and join existing ones using an invite code.
2. Each challenge should define rules based on predefined fitness metrics (e.g., steps, protein intake, sugar limit).
3. Participants should be able to submit daily metric values for the challenges they have joined.
4. The system should automatically evaluate submissions and assign points based on challenge rules.
5. The system should track penalties for missed submissions and accumulate fines per participant.
6. The system should maintain and display a real-time leaderboard for each challenge.
	    

---

# 3. Non-Goals (MVP Scope)

1. A native mobile application will not be developed for the MVP. The system will initially be delivered as a web application.
    
2. Only private challenges will be supported. Users can join challenges through invite codes, and public challenge discovery will not be available.
    
3. Advanced analytics or insights on user performance or submissions will not be included in the MVP.
    
4. Social features such as activity feeds, comments, or social media-style interactions will not be implemented in the MVP.
    

---

# 4. System Overview

The system provides a web-based platform where authenticated users can participate in fitness challenges. After logging in, users can view the list of challenges they have joined, explore challenge details, and see the current leaderboard.

Participants can submit their daily fitness metrics, which are evaluated by the backend according to the rules defined for that challenge.

The backend service processes submissions, calculates points and penalties, and updates the leaderboard state stored in the database. Challenge hosts can create new challenges and define rules based on predefined fitness metrics.

All system data, including users, challenges, submissions, and scores, are stored and managed in a PostgreSQL database.

---

# 5. High-Level Architecture

The system consists of three primary components:

- Web Frontend (React)
    
- Backend API (Go Service)
    
- PostgreSQL Database
    

Architecture flow:

User → Web Frontend → Backend API → PostgreSQL Database

The backend API handles authentication, challenge management, submission processing, scoring logic, and leaderboard generation.

**Backend**

```
Layered Architecture

Handler → Service → Repository
```

Backend Stack

```
Language: Go
HTTP: net/http
Router: chi (optional but recommended)
Database: PostgreSQL
Query Layer: sqlc + pgx
Migrations: golang-migrate
Auth: JWT + Refresh Tokens
Deployment: Docker
```
---

# 6. Database Design

The system uses a relational database (PostgreSQL) with normalized tables.

### Core Tables

**users**

Stores user accounts.

Fields:
- id
- name
- email
- password_hash
- created_at

---

**metrics**

Defines supported fitness metrics.

Examples:

- steps
- protein
- sugar
- water intake

Fields:

- id
- name
- unit
- created_at

---

**challenges**

Defines fitness challenges.

Fields:

- id
- name
- invite_code
- start_date
- end_date
- fine_amount
- status
- created_by
- created_at
- updated_at

---

### Relationship Tables

**user_challenges**

Represents user participation in challenges.

Fields:

- id
- user_id
- challenge_id
- role (host / cohost / participant)
- status (active / left)
- joined_at

Constraint:

- UNIQUE(user_id, challenge_id)

---

**challenge_metrics**

Defines scoring rules for each challenge.

Fields:

- id
- challenge_id
- metric_id
- target_value
- metric_type (min / max)
- points_awarded


---

### Event Tables

**daily_submissions**

Stores daily activity records.

Fields:

- id
- user_challenge_id
- date
- submission_type (submitted / missed)
- total_points
- total_fine
- video_url
- created_at

Constraint:

UNIQUE(user_challenge_id, date)

---

**submission_metric_values**

Stores evaluated metric values per submission.

Fields:

- id
- submission_id
- challenge_metric_id
- value
- passed
- points_awarded
- created_at

---

### Derived State

**challenge_scores**

Maintains leaderboard state.

Fields:

- id
- user_challenge_id
- total_points
- total_fines
- last_submission_date
- updated_at

Leaderboard ranking is computed at read time using SQL window functions.

---

### Authentication

**refresh_tokens**

Manages login sessions per device.

Fields:

- id
- user_id
- device_id
- token_hash
- ip_address
- user_agent
- expires_at
- revoked
- created_at
    

Refresh tokens are stored as hashed values for security.

![[FitnessChallengeArchitecture|800x400]]

---

# 7. Daily Submission Workflow

1. User submits daily metrics via the frontend.
2. Frontend sends request to backend. `POST /submissions`

3. Backend authenticates the user using the access token.
4. Backend validates:
    - user participation in challenge
    - challenge status is active
    - submission not already made for the day.
5. Backend starts a database transaction.
6. Insert new row in daily_submissions.
7. Load challenge metrics.
8. Evaluate each metric rule.
9. Insert rows into submission_metric_values.
10. Calculate total points.
11. Update challenge_scores.
12. Commit transaction.


```
User
 │
 │ Submit metrics
 ▼
Frontend
 │
 │ POST /submissions
 ▼
Backend
 │
 │ Validate JWT
 │ Validate user_challenge
 │ Check submission already exists
 │
 │ Begin Transaction
 │
 │ Insert daily_submissions
 │
 │ Fetch challenge_metrics
 │
 │ Evaluate rules
 │
 │ Insert submission_metric_values
 │
 │ Update challenge_scores
 │
 │ Commit
 │
 ▼
PostgreSQL
```

---

# 8. Missed Submission Handling

A scheduled background job runs once per day.

Steps:

1. Identify active challenges.
2. For each participant:
    - Check if submission exists for previous day.
3. If missing:
    - Insert a daily_submissions record with submission_type = missed.
    - Apply fine to challenge_scores.

This ensures full historical tracking of missed submissions.

---

# 9. Security Design

The system uses secure authentication practices:

- Passwords hashed using bcrypt or argon2.
- Access tokens implemented using JWT with short expiration.
- Refresh tokens are cryptographically random and stored as hashed values.
- Refresh tokens are issued per device session.
- HTTP-only cookies are used to store authentication tokens.

This design reduces exposure to XSS and token theft.

---

# 10. Scaling Considerations

Leaderboard queries are expected to have the highest read traffic.
To support this efficiently:
- Aggregated leaderboard state is stored in `challenge_scores`.
- Expensive aggregation queries on submission history are avoided.
- Backend APIs remain stateless to support horizontal scaling.
- Database indexing is applied on frequently queried fields:
    - challenge_id
    - user_challenge_id
    - date
- Scheduled background jobs handle fine calculation instead of request-time evaluation.

This architecture supports growth from small groups to thousands of participants.


```
                ┌─────────────────────┐
                │        User         │
                │  (Web Browser)     │
                └─────────┬───────────┘
                          │
                          ▼
                ┌─────────────────────┐
                │   Frontend (React)  │
                │   Web Application   │
                └─────────┬───────────┘
                          │ REST API
                          ▼
                ┌─────────────────────┐
                │   Backend API       │
                │      (Go)           │
                │                     │
                │ Auth               │
                │ Challenge Engine   │
                │ Scoring Logic      │
                └─────────┬───────────┘
                          │
                          ▼
                ┌─────────────────────┐
                │   PostgreSQL DB     │
                │                     │
                │ Users              │
                │ Challenges         │
                │ Submissions        │
                │ Scores             │
                └─────────────────────┘
```

System Context Diagram
```
User
 │
 ▼
Fitness Challenge Platform
 │
 ├── Email Notifications (future)
 ├── Video Storage (future)
 └── Mobile App (future)
```

**Database Schema Design**

```
users
  │
  │ 1
  │
  ▼
user_challenges
  │
  ├───────────────► challenges
  │
  ├───────────────► challenge_scores
  │
  └───────────────► daily_submissions
                       │
                       ▼
              submission_metric_values
                       │
                       ▼
               challenge_metrics
                       │
                       ▼
                    metrics


users
  │
  ▼
refresh_tokens
```


## Workflows
#### 1. Daily Submission Workflow

```
User
 │
 │ Submit metrics
 ▼
Frontend
 │
 │ POST /submissions
 ▼
Backend API
 │
 │ Validate user participation
 │ Validate challenge active
 │ Check submission uniqueness
 │
 │ Begin Transaction
 │
 │ Insert daily_submissions
 │
 │ Load challenge_metrics
 │ Evaluate metric rules
 │ Insert submission_metric_values
 │
 │ Update challenge_scores
 │
 │ Commit Transaction
 │
 ▼
PostgreSQL
```


#### 2. Missed Submission / Fine Calculation

```
Cron Scheduler
      │
      ▼
Backend Job
      │
      │ Fetch active challenges
      ▼
PostgreSQL
      │
      │ Find participants with missed submissions
      ▼
Backend Job
      │
      │ Apply fines
      │ Update challenge_scores
      ▼
PostgreSQL
```


### 3. Leaderboard Read Flow

```
User
 │
 │ View Leaderboard
 ▼
Frontend
 │
 │ GET /challenges/{id}/leaderboard
 ▼
Backend
 │
 │ Query challenge_scores
 │
 ▼
PostgreSQL
 │
 │ SELECT total_points
 │ FROM challenge_scores
 │ WHERE challenge_id = ?
 │
 ▼
Backend
 │
 │ Rank users using SQL window function
 │
 ▼
Frontend
 │
 │ Display leaderboard
 ▼
User
```

### 4. Cron Job – Missed Submission & Fine Processing

```
Cron Scheduler
 │
 │ Runs daily at 00:05
 ▼
Backend Job
 │
 │ Fetch active challenges
 ▼
PostgreSQL
 │
 │ SELECT challenges WHERE status='active'
 ▼
Backend Job
 │
 │ Find users with missing submission yesterday
 ▼
PostgreSQL
 │
 │ Query daily_submissions
 ▼
Backend Job
 │
 │ Insert missed submission
 │
 │ submission_type = missed
 │
 │ Apply fine
 │ total_fines += fine_amount
 ▼
PostgreSQL
 │
 │ Update challenge_scores
 ▼
Backend Job
 │
 │ Job completed
```