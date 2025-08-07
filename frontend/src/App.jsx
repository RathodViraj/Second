import { Routes, Route } from "react-router-dom";
import Layout from "./components/layout";
import Home from "./pages/Home.jsx";
import AddDocument from "./pages/AddDocument.jsx";
import Search from "./pages/Search.jsx";
import DocumentView from "./pages/DocumentView.jsx"; 
import ThemeToggle from "./components/ThemeToggle.jsx"; 
import Trending from "./pages/Trending.jsx";

const App = () => {
  return (    
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Home />} />
        <Route path="/search" element={<Search />} />
        <Route path="/add" element={<AddDocument />} />
        <Route path="/document/:id" element={<DocumentView />} />
        <Route path="/trending" element={<Trending />} />
      </Route>
    </Routes>
  );
};

export default App;
