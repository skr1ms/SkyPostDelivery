import { Outlet, Link, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import './Layout.css';
import logo from '../../assets/icon/logo.jpg';

const Layout = () => {
  const location = useLocation();
  const { user, logout } = useAuth();

  const navItems = [
    { path: '/dashboard', label: 'Dashboard', icon: '📊' },
    { path: '/drones', label: 'Дроны', icon: '🚁' },
    { path: '/goods', label: 'Товары', icon: '📦' },
    { path: '/parcel-automats', label: 'Постаматы', icon: '🏪' },
    { path: '/monitoring', label: 'Мониторинг', icon: '📈' },
  ];

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="logo">
          <img src={logo} alt="SkyPost Logo" className="logo-image" />
          <h1>SkyPost Delivery Admin</h1>
        </div>
        <nav className="nav">
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`nav-item ${location.pathname === item.path ? 'active' : ''}`}
            >
              <span className="nav-icon">{item.icon}</span>
              <span className="nav-label">{item.label}</span>
            </Link>
          ))}
        </nav>
        <div className="sidebar-footer">
          <div className="user-info">
            <div className="user-avatar">
              {user?.full_name.charAt(0).toUpperCase()}
            </div>
            <div className="user-details">
              <div className="user-name">{user?.full_name}</div>
              <div className="user-role">{user?.role}</div>
            </div>
          </div>
          <button className="logout-button" onClick={logout}>
            🚪 Выйти
          </button>
        </div>
      </aside>
      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
};

export default Layout;
