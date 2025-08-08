import { Link } from "react-router-dom";
import ThemeToggle from "./ThemeToggle.jsx";
import "./Navbar.css";

export default function Navbar() {
  return (
    <nav className="navbar">
      <div className="navbar-left">
        <Link to="/" className="navbar-logo">DocSearch</Link>
        <ul className="navbar-links">
          <li><Link to="/">Home</Link></li>
          <li><Link to="/add">Add Document</Link></li>
          <li><Link to="/trending">Trending🔥</Link></li>
        </ul>
      </div>
      <div className="navbar-right">
        <ThemeToggle />
      </div>
    </nav>
  );
}
