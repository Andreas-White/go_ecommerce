import './Footer.css';

export default function Footer() {
  return (
    <footer className="footer">
      <div className="footer-copyright">&copy; {new Date().getFullYear()} SnapCart. All rights reserved.</div>
      <div className="footer-links">
        <a href="/privacy" className="footer-link">Privacy Policy</a>
        <a href="/terms" className="footer-link">Terms of Service</a>
      </div>
    </footer>
  );
} 