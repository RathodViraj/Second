import { useParams } from "react-router-dom";
import { useEffect, useState } from "react";

const DocumentView = () => {
  const { id } = useParams();
  const [doc, setDoc] = useState(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchDoc = async () => {
      try {
        const res = await fetch(`http://localhost:8080/document/${id}`);
        if (!res.ok) throw new Error("Failed to fetch document");
        const data = await res.json();
        setDoc(data);
      } catch (err) {
        setError(err.message);
      }
    };
    fetchDoc();
  }, [id]);

  if (error) return <p>Error: {error}</p>;
  if (!doc) return <p>Loading...</p>;

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-4">{doc.title}</h1>
      <p className="mb-2 text-gray-700 whitespace-pre-line">{doc.content}</p>
      {doc.tags?.length > 0 && (
        <p className="text-sm text-gray-500 mt-4">Tags: {doc.tags.join(", ")}</p>
      )}
    </div>
  );
};

export default DocumentView;
