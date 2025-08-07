import React, { useEffect, useState, useRef } from "react";
import { useNavigate } from "react-router-dom";

const Home = () => {
  const [query, setQuery] = useState("");
  const [suggestions, setSuggestions] = useState([]);
  const wsRef = useRef(null);
  const navigate = useNavigate();

  useEffect(() => {
    // Connect to WebSocket on mount
    wsRef.current = new WebSocket("ws://localhost:8080/typeahead");

    wsRef.current.onopen = () => {
      console.log("WebSocket connected");
    };

    wsRef.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.suggestions) {
        setSuggestions(data.suggestions);
      }
    };

    wsRef.current.onerror = (err) => {
      console.error("WebSocket error:", err);
    };

    wsRef.current.onclose = () => {
      console.log("WebSocket disconnected");
    };

    return () => {
      wsRef.current?.close();
    };
  }, []);

  // Send typeahead request on input change
  useEffect(() => {
    const handler = setTimeout(() => {
      if (
        query.length >= 3 &&
        query.length <= 20 &&
        wsRef.current?.readyState === WebSocket.OPEN
      ) {
        const payload = {
          prefix: query,
          limit: 5,
        };
        
        console.log("Sending typeahead request:", query);
        
        wsRef.current.send(JSON.stringify(payload));
      } else {
        setSuggestions([]);
      }
    }, 300); // 300ms debounce

    return () => clearTimeout(handler); // cleanup previous timer
  }, [query]);


  const handleSearch = () => {
    if (query.trim() !== "") {
      navigate(`/search?q=${encodeURIComponent(query)}`);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === "Enter") handleSearch();
  };

  const handleSuggestionClick = (s) => {
    setQuery(s);
    setSuggestions([]);
  };

  return (
    <div
      style={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        flexDirection: "column",
        marginTop: "4rem", // ✅ gives some space under the navbar
      }}
    >
      <div style={{ display: "flex", flexDirection: "column", alignItems: "center" }}>
        <div style={{ display: "flex", gap: "0.5rem" }}>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search..."
            style={{
              padding: "0.75rem 1.5rem",
              borderRadius: "999px",
              border: "1px solid #ccc",
              width: "300px",
              fontSize: "1rem",
            }}
          />
          <button
            onClick={handleSearch}
            style={{
              padding: "0.75rem 1.5rem",
              borderRadius: "999px",
              border: "none",
              backgroundColor: "#007bff",
              color: "#fff",
              fontSize: "1rem",
              cursor: "pointer",
            }}
          >
            Search
          </button>
        </div>

        {/* Typeahead Suggestions */}
        {suggestions.length > 0 && (
          <ul
            style={{
              marginTop: "0.5rem",
              backgroundColor: "#fff",
              border: "1px solid #ccc",
              borderRadius: "0.5rem",
              width: "300px",
              listStyle: "none",
              padding: "0.5rem",
              boxShadow: "0px 4px 8px rgba(0,0,0,0.1)",
              zIndex: 10,
              position: "absolute",
              top: "calc(50% + 2.5rem)",
              color: "black",
            }}
          >
            {suggestions.map((s, index) => (
              <li
                key={index}
                onClick={() => handleSuggestionClick(s)}
                style={{
                  padding: "0.5rem",
                  cursor: "pointer",
                }}
              >
                {s}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};

export default Home;
