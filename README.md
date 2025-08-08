
# Document Search API

This project is a backend service for storing and searching documents with features like:
- **Indexed search** for fast lookups
- **Type-ahead (autocomplete)** for search suggestions
- **Rate limiting** to prevent abuse
- **View tracking** using Redis for ranking documents

The service uses:
- **Go** for backend logic
- **MongoDB** for document storage
- **Redis** for caching, scoring, and rate limiting

---

## Features & Concepts

### 1. Index Search
**What it is:**  
Index search uses a pre-built data structure to locate documents quickly instead of scanning every document.  
MongoDB creates **indexes** on fields (like `title` or `content`) so lookups are extremely fast.

**How it works here:**
- We create a **text index** in MongoDB on relevant fields (`title`, `content`).
- When you search for a word, MongoDB matches only the indexed terms, skipping irrelevant documents.
- This is much faster than a full collection scan.

**Example:**
```js
db.documents.createIndex({ title: "text", content: "text" });
````

When searching `"network security"`, MongoDB instantly finds documents containing these terms using the index.

---

### 2. Type-ahead Search (Autocomplete)

**What it is:**
Type-ahead shows possible matches as you start typing (e.g., typing "net" suggests "network security", "network programming", etc.).

**How it works here:**

* We maintain a sorted set or prefix list in Redis containing all searchable terms.
* As a user types, we fetch terms starting with that prefix.
* This returns suggestions without querying the whole database.

**Example:**
If Redis stores:

```
network security
network programming
netflix data analysis
```

Typing `"net"` immediately shows all terms starting with `"net"`.

---

### 3. Rate Limiter

**What it is:**
Rate limiting controls how many requests a user can make within a time window to avoid abuse (e.g., preventing one IP from making thousands of searches in seconds).

**How it works here:**

* We use Redis to store a counter for each IP or user.
* When a request comes in:

  1. Increment the counter in Redis.
  2. If the counter exceeds the allowed limit, block the request.
  3. The counter resets after the time window (e.g., 60 seconds).

**Example:**

```
Limit: 10 searches per minute
User searches 12 times in 60 seconds → last 2 requests are rejected
```

---

## How the System Works Together

1. **Document Insertion**

   * Documents are stored in MongoDB.
   * Title & content are indexed for search.
   * Keywords are stored in Redis for type-ahead.

2. **Searching**

   * User sends a query.
   * Rate limiter checks if the request is allowed.
   * MongoDB uses its text index to quickly return matching documents.
   * If using type-ahead mode, Redis returns suggestions instantly.

3. **Document Viewing**

   * When a document is clicked, Redis increments its score in a sorted set (`document_scores`).
   * This score can be used to show trending/popular documents.

4. **Performance**

   * **MongoDB** handles main document storage & full-text search.
   * **Redis** handles quick prefix lookups, trending docs, and rate limiting.

---

## Example Request Flow

1. User starts typing `"net"`.

   * **Redis** returns `"network security"`, `"network programming"`, `"netflix data analysis"`.

2. User selects `"network security"`.

   * **MongoDB** returns all matching documents using the text index.

3. User clicks a document.

   * **Redis** increments its popularity score.

4. Another user searches the same.

   * The popular document appears higher in results.

---

## Tech Stack

* **Go** — Backend service
* **MongoDB** — Document database
* **Redis** — Caching, type-ahead, rate limiter, and scoring
* **React** - Frontend

---

## Example Commands

### Create Text Index in MongoDB

```bash
db.documents.createIndex({ title: "text", content: "text" });
```

### Redis Commands for Type-ahead

```bash
ZADD search_terms 0 "network security"
ZADD search_terms 0 "network programming"
ZRANGEBYLEX search_terms [net [net\xff
```

### Redis Rate Limiter (per IP)

```bash
INCR user:192.168.1.5
EXPIRE user:192.168.1.5 60
```


