import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { dronesAPI, goodsAPI, parcelAutomatsAPI } from '../../api';
import './Dashboard.css';

interface Stats {
  drones: number;
  goods: number;
  parcelAutomats: number;
}

const Dashboard = () => {
  const navigate = useNavigate();
  const [stats, setStats] = useState<Stats>({ drones: 0, goods: 0, parcelAutomats: 0 });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      setLoading(true);
      const [dronesRes, goodsRes, automatsRes] = await Promise.all([
        dronesAPI.getAll(),
        goodsAPI.getAll(),
        parcelAutomatsAPI.getAll(),
      ]);

      setStats({
        drones: dronesRes.data.length,
        goods: goodsRes.data.length,
        parcelAutomats: automatsRes.data.length,
      });
    } catch (error) {
      // Silent
    } finally {
      setLoading(false);
    }
  };

  const cards = [
    { title: 'Дроны', value: stats.drones, icon: '', color: '#6c5ce7', path: '/drones' },
    { title: 'Товары', value: stats.goods, icon: '📦', color: '#ff6b9d', path: '/goods' },
    { title: 'Постаматы', value: stats.parcelAutomats, icon: '🏪', color: '#00d4aa', path: '/parcel-automats' },
  ];

  return (
    <div className="dashboard">
      <div className="page-header">
        <h1>Dashboard</h1>
        <p>Общая статистика системы</p>
      </div>

      <div className="stats-grid">
        {cards.map((card) => (
          <div
            key={card.title}
            className="stat-card"
            style={{ '--card-color': card.color } as React.CSSProperties}
            onClick={() => navigate(card.path)}
          >
            <div className="stat-icon">{card.icon}</div>
            <div className="stat-info">
              <h3>{card.title}</h3>
              <div className="stat-value">
                {loading ? '...' : card.value}
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="info-section">
        <div className="info-card">
          <h2>Добро пожаловать в Admin Panel</h2>
          <p>
            Используйте боковое меню для управления дронами, товарами и постаматами.
          </p>
          <ul>
            <li><strong>Дроны</strong> - управление парком дронов</li>
            <li><strong>Товары</strong> - добавление и редактирование товаров</li>
            <li><strong>Постаматы</strong> - настройка пунктов выдачи</li>
            <li><strong>Мониторинг</strong> - просмотр метрик и логов системы</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
