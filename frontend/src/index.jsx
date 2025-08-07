import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.jsx";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import AddDocument from "./components/AddDocument.jsx";
import Search from "./components/Search.jsx";
import "./index.css";

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/add" element={<AddDocument />} />
        <Route path="/search" element={<Search />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);
