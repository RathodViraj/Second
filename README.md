# Backend Search & Performance Optimization System

## Overview
This project implements a **high-performance backend search system** with:
- **Index-based search** for fast lookups
- **Type-ahead (autocomplete)** for real-time predictive search
- **Rate limiting** to prevent abuse and ensure fair API usage
- **Trending documents tracker** to highlight popular content

**Tech Stack:** Go, MongoDB, Redis, Docker, REST API

---

## Features

### 1. Index Search
**What it is:**  
Index search uses database indexes (in MongoDB) to quickly find matching records without scanning the entire collection.  
By indexing fields like `name` or `keywords`, the database can locate matches in milliseconds, even in datasets with millions of documents.

**How it works in this project:**  
- Created MongoDB indexes on searchable fields.
- Queries use `$regex` and `$text` search with index optimization.
- Significantly reduces query execution time.

---

### 2. Type-Ahead Search (Autocomplete)
**What it is:**  
A type-ahead search predicts and suggests results as the user types, improving the search experience.

**How it works in this project:**  
- Captures the user’s input in real-time.
- Performs partial matches against indexed fields.
- Returns the top N relevant results for each keystroke.
- Uses MongoDB's efficient index lookups for low latency.

---

### 3. Rate Limiter
**What it is:**  
A rate limiter controls the number of API requests a client can make in a given time frame, preventing abuse and ensuring fair usage.

**How it works in this project:**  
- Implemented using Redis as a fast in-memory store.
- Stores request counts per user/IP with an expiration time.
- If the request limit is exceeded, the API returns an error with a cooldown period.

---

### 4. Trending Documents Tracker
**What it is:**  
Tracks the most accessed documents and ranks them in real time.

**How it works in this project:**  
- Uses Redis Sorted Sets to store document IDs with their access frequency.
- On every document view, increments its score.
- A separate endpoint retrieves the top N trending documents.

---

## System Workflow

1. **User Search Request**  
   - The API receives the search query.
   - Type-ahead logic processes partial inputs.
   - Index search is performed on MongoDB for fast matching.

2. **Rate Limiting Check**  
   - Redis checks if the user/IP has exceeded allowed request limits.
   - If yes → request is blocked.  
   - If no → search proceeds.

3. **Return Results**  
   - Matching documents are returned.
   - Document access is logged in Redis for trending tracking.

4. **Trending Update**  
   - Each access increments the document’s popularity score in Redis.
   - Top trending documents are available via a dedicated API.

---

## Example API Endpoints

| Endpoint                | Method | Description                              |
|------------------------|--------|------------------------------------------|
| `/search?q=query`      | GET    | Search documents (type-ahead + index)    |
| `/document/{id}`       | GET    | Retrieve a single document               |
| `/trending`            | GET    | Get top trending documents               |

---

## Why This Matters
- **Performance:** Index search and Redis caching make lookups extremely fast.
- **Scalability:** Handles high traffic without degrading performance.
- **User Experience:** Type-ahead search improves engagement.
- **Security & Fairness:** Rate limiting prevents abuse and server overload.

---
