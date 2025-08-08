import React, { useEffect, useState } from "react";

const Trending = () => {
  const [trendingDocs, setTrendingDocs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchTrendingDocs = async () => {
      try {
        const response = await fetch("http://localhost:8080/trending");
        if (!response.ok) throw new Error("Failed to fetch trending documents");
        const data = await response.json();
        setTrendingDocs(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchTrendingDocs();
  }, []);

  if (loading) return <div className="p-4 text-lg">Loading...</div>;
  if (error) return <div className="p-4 text-red-600">Error: {error}</div>;

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <h1 className="text-3xl font-semibold mb-6">🔥 Trending Documents (Last 1 Hour)</h1>
      <ul className="space-y-3">
        {Array.isArray(trendingDocs) && trendingDocs.length > 0 ?  (
          trendingDocs.slice(0, 50).map((doc, index) => (
            <li key={index} className="border-b pb-2 flex justify-between items-center">
              <span className="text-lg font-medium">{doc.title}</span>
              <span className="text-sm text-gray-600">{doc.views} views</span>
            </li>
          ))
        )
          : <li className="text-gray-500">No trending documents found.</li>
        }
      </ul>
    </div>
  );
};

export default Trending;
