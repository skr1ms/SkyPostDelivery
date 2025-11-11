import './MonitoringPage.css';
import { MONITORING_CONFIG } from '../../config/api_config';

const MonitoringPage = () => {
  const { minioURL, grafanaURL, rabbitmqURL } = MONITORING_CONFIG;

  const openExternal = (url: string) => {
    window.open(url, '_blank', 'noopener,noreferrer');
  };

  return (
    <div className="monitoring-page">
      <div className="page-header">
        <div>
          <h1>📈 Мониторинг Системы</h1>
          <p>Просмотр метрик, логов и хранилища</p>
        </div>
      </div>

      <div className="monitoring-grid">
        <div className="monitoring-card minio-card">
          <div className="card-icon">🗄️</div>
          <h2>MinIO Storage</h2>
          <p>Объектное хранилище для QR-кодов и файлов системы</p>
          <button className="open-button minio-button" onClick={() => openExternal(minioURL)}>
            🚀 Открыть MinIO Console
          </button>
        </div>

        <div className="monitoring-card grafana-card">
          <div className="card-icon">📊</div>
          <h2>Grafana Dashboard</h2>
          <p>Визуализация метрик и логов в реальном времени</p>
          <button className="open-button grafana-button" onClick={() => openExternal(grafanaURL)}>
            🚀 Открыть Grafana
          </button>
        </div>

        <div className="monitoring-card rabbitmq-card">
          <div className="card-icon">🐰</div>
          <h2>RabbitMQ Management</h2>
          <p>Управление очередями сообщений и мониторинг брокера</p>
          <button className="open-button rabbitmq-button" onClick={() => openExternal(rabbitmqURL)}>
            🚀 Открыть RabbitMQ
          </button>
        </div>
      </div>

      <div className="info-section">
        <h2>ℹ️ Информация о Мониторинге</h2>
        
        <div className="info-cards">
          <div className="info-card">
            <h3>🗄️ MinIO</h3>
            <ul>
              <li>Хранение QR-кодов пользователей</li>
              <li>Бэкапы базы данных</li>
              <li>Логи системы</li>
              <li>Статические файлы</li>
            </ul>
          </div>

          <div className="info-card">
            <h3>📊 Grafana</h3>
            <ul>
              <li>Метрики производительности (CPU, RAM, Disk)</li>
              <li>Количество запросов к API</li>
              <li>Статус дронов в реальном времени</li>
              <li>Ошибки и предупреждения</li>
              <li>Логи всех сервисов</li>
            </ul>
          </div>

          <div className="info-card">
            <h3>🐰 RabbitMQ</h3>
            <ul>
              <li>Управление очередями доставок</li>
              <li>Мониторинг consumers и connections</li>
              <li>Статистика обработки сообщений</li>
              <li>Dead Letter Queue (DLQ)</li>
              <li>Метрики памяти и производительности</li>
            </ul>
          </div>
          
        </div>
      </div>
    </div>
  );
};

export default MonitoringPage;
