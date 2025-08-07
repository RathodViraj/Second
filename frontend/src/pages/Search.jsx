import { useEffect, useState } from "react";
import { useLocation } from "react-router-dom";

const Search = () => {
  const location = useLocation();
  const queryParams = new URLSearchParams(location.search);
  const query = queryParams.get("q");

  const [results, setResults] = useState([]);

  useEffect(() => {
    if (!query) return;

    const timer = setTimeout(() => {
      const fetchResults = async () => {
        try {
          const res = await fetch(`http://localhost:8080/search?q=${encodeURIComponent(query)}`);
          const data = await res.json();
          setResults(data.results || []);
        } catch (err) {
          console.error("Search failed:", err);
        }
      };

      fetchResults();
    }, 300); // 300ms debounce

    return () => clearTimeout(timer);
  }, [query]);

  return (
    <div className="p-4">
      <h2 className="text-xl mb-4">Search Results for: {query}</h2>
      <ul className="list-disc pl-5 space-y-2">
        {results.map((doc, index) => (
          <li key={doc.id}>
            <Link to={`/document/${doc.id}`} className="text-blue-500 hover:underline">
              {doc.title}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default Search;
