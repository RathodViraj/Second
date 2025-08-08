import React, { useEffect, useState, useRef } from "react";
import { useNavigate } from "react-router-dom";

const Home = () => {
  const [query, setQuery] = useState("");
  const [suggestions, setSuggestions] = useState([]);
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [manualSelection, setManualSelection] = useState(false);
  const wsRef = useRef(null);
  const navigate = useNavigate();

  useEffect(() => {
    wsRef.current = new WebSocket("ws://localhost:8080/typeahead");

    wsRef.current.onopen = () => {
      console.log("WebSocket connected");
    };

    wsRef.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.suggestions) {
        setSuggestions(data.suggestions);
        setSelectedIndex(-1); // reset selection on new suggestions
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

  useEffect(() => {
    const handler = setTimeout(() => {
      if (manualSelection) {
        setManualSelection(false);
        return;
      }

      if (
        query.length >= 3 &&
        query.length <= 20 &&
        wsRef.current?.readyState === WebSocket.OPEN
      ) {
        const payload = {
          prefix: query,
          limit: 5,
        };

        wsRef.current.send(JSON.stringify(payload));
      } else {
        setSuggestions([]);
        setSelectedIndex(-1);
      }
    }, 400);

    return () => clearTimeout(handler);
  }, [query]);

  const handleSearch = () => {
    if (query.trim() !== "") {
      navigate(`/search?q=${encodeURIComponent(query)}`);
    }
  };

  const handleKeyDown = (e) => {
    if (suggestions.length > 0) {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        const newIndex = (selectedIndex + 1) % suggestions.length;
        setSelectedIndex(newIndex);
        setQuery(suggestions[newIndex]);
        setManualSelection(true);
        return;
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        const newIndex =
          (selectedIndex - 1 + suggestions.length) % suggestions.length;
        setSelectedIndex(newIndex);
        setQuery(suggestions[newIndex]);
        setManualSelection(true);
        return;
      }
    }

    if (e.key === "Enter") {
      handleSearch();
    }
  };


  const handleSuggestionClick = (s) => {
    setQuery(s);
    setSuggestions([]);
    setSelectedIndex(-1);
    setManualSelection(true)
  };

  return (
    <div
      style={{
        display: "flex",
        height: "100vh",
        justifyContent: "center",
        alignItems: "center",
        flexDirection: "column",
        position: "relative",
        top: "-100px",
      }}
    >
      <div style={{ display: "flex", flexDirection: "column", alignItems: "center" }}>
        <div style={{ display: "flex", gap: "0.5rem" }}>
          <input
            type="text"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setSelectedIndex(-1);
            }}
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
                  backgroundColor: selectedIndex === index ? "#f0f0f0" : "transparent",
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
