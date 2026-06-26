# Document Search & Management System

A full-stack web application for managing, searching, and discovering documents with real-time trending analytics and intelligent search capabilities.

---

## Project Overview

This application provides a complete document management solution with:
- **Document Management**: Add and organize documents with titles, content, and tags
- **Advanced Search**: Full-text search with stopword filtering and tokenization
- **Trending Analytics**: Real-time trending documents using sliding window algorithm
- **Typeahead/Autocomplete**: Live search suggestions via WebSocket
- **Rate Limiting**: API protection using Leaky Bucket algorithm
- **Caching**: High-performance caching with Redis

---

## How It Works

### 1. **Document Management**

#### Adding Documents
- User navigates to `/add` page in the frontend
- Fills in form with title, content, and tags
- Frontend sends POST request to `/add` endpoint
- Backend validates required fields (title, content)
- Document is stored in MongoDB with timestamps
- User receives confirmation and is redirected

**Data Model:**
```
{
  "id": ObjectID,
  "title": string,
  "content": string,
  "tags": [string],
  "created": timestamp
}
```

### 2. **Search Functionality**

#### How Search Works
1. User enters query on `/search` page
2. Frontend sends GET request with query parameter to `/search` endpoint
3. Backend processes the search:
   - **Tokenization**: Query is split into individual words
   - **Stopword Removal**: Common words (the, a, and, etc.) from `stopwords.txt` are filtered out
   - **Full-Text Search**: MongoDB searches documents by title, content, and tags
   - **Caching**: Results are cached in Redis for performance
4. Results displayed in real-time with highlighting

### 3. **Trending Documents**

#### Sliding Window Algorithm
The application tracks document views using a **sliding window** approach:

**How it works:**
- Each minute window, the system records how many times documents were viewed
- Trending documents are determined by view count over the current minute window
- Sliding window automatically advances every minute
- Top documents by view frequency are cached in Redis
- Frontend fetches trending docs from `/trending` endpoint

**Backend Process:**
- `startSlidingWindow()` goroutine runs continuously
- Every minute, `repo.SlideWindow()` is called
- `waitUntilNextMinute()` ensures synchronization with clock
- Trending data is updated and stored in Redis

### 4. **Typeahead/Autocomplete**

#### Real-Time Search Suggestions
- WebSocket connection established at `/typeahead` endpoint
- As user types in search box, suggestions are streamed live
- Uses tokenizer to process partial queries
- Redis stores indexed terms for fast lookup
- Live suggestions without page refresh

### 5. **Rate Limiting**

#### Leaky Bucket Algorithm
The rate limiter protects API endpoints from abuse.

**Algorithm Used:** **Leaky Bucket**
- Implemented as a Redis Lua script (`leaky_bucket.lua`)
- Each client IP has a "bucket" with fixed capacity (default: 5 tokens)
- Tokens "leak" at a constant rate (default: 0.5 tokens/second)
- New requests consume 1 token
- If bucket is full → request rejected (429 Too Many Requests)
- If bucket has space → request allowed, token consumed

**Configuration (in main.go):**
```go
middleware.NewRateLimiter(rdb, 0.5, 5)
                              ↓    ↓
                           rate  capacity
```

**Benefits:**
- Smooth traffic shaping
- Fair rate limiting per IP
- Redis-based (distributed)
- No request bursts

### 6. **Caching Strategy**

Redis is used for multiple purposes:
- **Search Results**: Cached for 5-10 minutes
- **Trending Data**: Updated every minute
- **Rate Limiting**: Per-IP state tracking
- **Typeahead Index**: Terms and suggestions
- **Session Data**: Temporary WebSocket data

---

## API Endpoints

### Documents

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/add` | Add a new document |
| GET | `/search?q=query` | Search documents |
| GET | `/document/:id` | Get document by ID |

### Trending & Analytics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trending` | Get trending documents |

### Real-Time

| Method | Endpoint | Description |
|--------|----------|-------------|
| WS | `/typeahead` | WebSocket for autocomplete suggestions |

---

## Middleware

### Rate Limiting Middleware
- Applied to all routes in Gin
- Checks client IP against rate limit
- Uses Redis Lua scripts for atomic operations
- Returns 429 status if limit exceeded

---


## Architecture

### Tech Stack

**Backend:**
- **Language**: Go (Golang)
- **Framework**: Gin Web Framework
- **Databases**: 
  - MongoDB (Document storage)
  - Redis (Caching, trending, rate limiting, typeahead)
- **Port**: 8080

**Frontend:**
- **Framework**: React 19
- **Build Tool**: Vite
- **Routing**: React Router v7
- **Port**: 5173

---

## Project Structure

```
backend/
├── main.go                    # Server entry point and route configuration
├── go.mod                     # Go dependencies
├── db/
│   ├── mongoDB.go            # MongoDB connection and setup
│   └── redis.go              # Redis connection and initialization
├── handler/                  # HTTP request handlers
│   ├── documentHandler.go    # Document CRUD operations
│   ├── trendingHandler.go    # Trending documents endpoint
│   ├── wsHandler.go          # WebSocket for typeahead
│   └── jsonError.go          # Error response formatting
├── middleware/
│   ├── rate_limiter.go       # Rate limiter middleware
│   └── leaky_bucket.lua      # Leaky Bucket algorithm (Redis Lua script)
├── model/
│   ├── document.go           # Document data structure
│   └── index_req.go          # Search request model
├── repository/               # Data access layer
│   ├── document.go           # Document repository
│   └── trending.go           # Trending repository
├── typeahead/
│   └── typeahead.go          # Typeahead search logic
└── utils/
    ├── tokenizer.go          # Text tokenization
    └── stopwords.txt         # Common English stopwords

frontend/
├── package.json              # Dependencies and scripts
├── vite.config.js            # Vite configuration
├── index.html                # Entry HTML file
├── src/
│   ├── main.jsx              # React app entry
│   ├── App.jsx               # Main app routing
│   ├── index.css             # Global styles
│   └── components/           # Reusable components
│       ├── Layout.jsx        # Main layout wrapper
│       ├── navbar.jsx        # Navigation bar
│       ├── navbar.css        # Navigation styles
│       └── ThemeToggle.jsx    # Dark/Light mode toggle
│   └── pages/                # Page components
│       ├── Home.jsx          # Landing page
│       ├── AddDocument.jsx   # Document creation form
│       ├── Search.jsx        # Search results page
│       ├── DocumentView.jsx  # Individual document viewer
│       └── Trending.jsx      # Trending documents page
└── public/
    └── index.html            # Static HTML
```

---

## Data Flow Diagram

```
User Browser (React App)
    ↓
Frontend Routes (/search, /add, /trending, /document/:id)
    ↓
Vite Dev Server (localhost:5173)
    ↓ (HTTP/WebSocket)
Gin Backend (localhost:8080)
    ↓
CORS Middleware → Rate Limiting Middleware → Handlers
    ↓
Database Layer:
  ├─ MongoDB (persistent storage)
  └─ Redis (cache, trending, rate limiting, typeahead)
```

---

## Key Features Explained

### Tokenization & Stopword Removal
- Breaks text into individual words
- Removes common words that don't add meaning
- Improves search accuracy and performance
- Uses `stopwords.txt` for English stopwords

### Document Indexing
- Automatic indexing on MongoDB
- Tags enable categorization
- Timestamps track document creation
- Enables filtering and sorting

### Real-Time Updates
- Trending data updates every minute
- WebSocket keeps typeahead suggestions current
- Redis pub/sub for broadcasting changes

---

## Getting Started

### Backend Setup

1. **Prerequisites:**
   - Go 1.18+
   - MongoDB running locally or remote
   - Redis running locally or remote

2. **Install Dependencies:**
   ```bash
   cd backend
   go mod download
   ```

3. **Configure Connections:**
   - Update MongoDB connection string in `db/mongoDB.go`
   - Update Redis connection in `db/redis.go`

4. **Run Backend:**
   ```bash
   go run main.go
   ```
   Server starts on `http://localhost:8080`

### Frontend Setup

1. **Prerequisites:**
   - Node.js 16+
   - npm or yarn

2. **Install Dependencies:**
   ```bash
   cd frontend
   npm install
   ```

3. **Run Development Server:**
   ```bash
   npm run dev
   ```
   App starts on `http://localhost:5173`

4. **Build for Production:**
   ```bash
   npm run build
   ```

---

**Common Status Codes:**
- `400`: Bad Request (missing/invalid fields)
- `404`: Document Not Found
- `429`: Rate Limit Exceeded
- `500`: Internal Server Error

---

## Frontend Features

- **Responsive Design**: Works on desktop and mobile
- **Dark Mode**: Theme toggle for user preference
- **Real-Time Search**: Typeahead suggestions as you type
- **Navigation**: Navbar with links to all sections
- **Document Viewer**: Displays full document with metadata
- **Trending Dashboard**: Shows popular documents

---

## Performance Optimizations

1. **Redis Caching**: Frequently accessed data cached in memory
2. **Lua Scripts**: Atomic operations for rate limiting
3. **Tokenization**: Reduces search complexity
4. **Stopword Filtering**: Minimizes index size
5. **Sliding Window**: Efficient trending calculation
6. **WebSocket**: Reduces overhead vs polling

---

## Development Workflow

1. Start Redis and MongoDB services
2. Run backend: `go run main.go`
3. In another terminal, run frontend: `npm run dev`
4. Open `http://localhost:5173` in browser
5. Test features and develop


