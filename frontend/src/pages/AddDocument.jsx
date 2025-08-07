import React, { useState } from "react";
import "./AddDocument.css";

const AddDocument = () => {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [tags, setTags] = useState("");

  const handleSubmit = async () => {
    try {
      const response = await fetch("http://localhost:8080/add", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title,
          content,
          tags: tags.split(",").map((tag) => tag.trim()).filter(Boolean),
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        alert(`Failed to add document: ${errorData.error || response.statusText}`);
        return;
      }

      const data = await response.json();
      alert(`Document added! ID: ${data.doc_id}`);
      setTitle("");
      setContent("");
      setTags("");
    } catch (error) {
      console.error("Error submitting document:", error);
      alert("Something went wrong. Please try again later.");
    }
  };

  return (
    <div className="add-doc-container">
      <h2>Add New Document</h2>
      <div className="form-group">
        <label htmlFor="title">Title</label>
        <input
          id="title"
          type="text"
          placeholder="Enter title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />
      </div>

      <div className="form-group">
        <label htmlFor="content">Content</label>
        <textarea
          id="content"
          placeholder="Write your document here..."
          value={content}
          onChange={(e) => setContent(e.target.value)}
        ></textarea>
      </div>

      <div className="form-group">
        <label htmlFor="tags">Tags (comma-separated)</label>
        <input
          id="tags"
          type="text"
          placeholder="e.g. go, backend, search"
          value={tags}
          onChange={(e) => setTags(e.target.value)}
        />
      </div>

      <button className="submit-btn" onClick={handleSubmit}>
        Submit Document
      </button>
    </div>
  );
};

export default AddDocument;
