import { useEffect, useState, useRef } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { dronesAPI } from "../../api";
import type { Drone } from "../../types";
import Button from "../../components/Button/Button";
import "./DroneDetailPage.css";

const DroneDetailPage = () => {
  const { droneId } = useParams<{ droneId: string }>();
  const navigate = useNavigate();
  const [drone, setDrone] = useState<Drone | null>(null);
  const [loading, setLoading] = useState(true);
  const [isVideoConnected, setIsVideoConnected] = useState(false);
  const [frameCount, setFrameCount] = useState(0);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const adminWsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!droneId) return;

    loadDrone();
    connectVideoStream();
    connectAdminWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (adminWsRef.current) {
        adminWsRef.current.close();
      }
    };
  }, [droneId]);

  const connectAdminWebSocket = () => {
    const ws = new WebSocket("ws://localhost:8081/ws/admin");
    adminWsRef.current = ws;

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === "drones_status" && data.drones && droneId) {
        const update = data.drones.find((d: any) => d.drone_id === droneId);
        if (update && drone) {
          setDrone({
            ...drone,
            battery_level: update.battery_level,
            status: update.status,
            current_delivery_id: update.current_delivery_id,
            speed: update.speed,
            error_message: update.error_message,
          });
        }
      }
    };

    ws.onclose = () => {
      setTimeout(connectAdminWebSocket, 3000);
    };
  };

  const loadDrone = async () => {
    if (!droneId) return;
    setLoading(true);
    try {
      const response = await dronesAPI.getById(droneId);
      setDrone(response.data);
    } catch (error) {
      // Silent
    } finally {
      setLoading(false);
    }
  };

  const connectVideoStream = () => {
    if (!droneId) return;

    const ws = new WebSocket(`ws://localhost:8081/ws/drone/${droneId}/video`);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsVideoConnected(true);
      setFrameCount(0);

      if (canvasRef.current) {
        const ctx = canvasRef.current.getContext("2d");
        if (ctx) {
          ctx.fillStyle = "red";
          ctx.fillRect(0, 0, 100, 100);
        }
      }
    };

    ws.onmessage = (event) => {
      const canvas = canvasRef.current;
      if (!canvas) {
        return;
      }

      const ctx = canvas.getContext("2d");
      if (!ctx) {
        return;
      }

      const base64Data = event.data;
      const binaryString = atob(base64Data);
      const bytes = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }
      const blob = new Blob([bytes], { type: "image/jpeg" });

      const img = new Image();

      img.onload = () => {
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.drawImage(img, 0, 0, canvas.width, canvas.height);

        URL.revokeObjectURL(img.src);

        setFrameCount((prev) => prev + 1);
      };

      img.onerror = () => {
        URL.revokeObjectURL(img.src);
      };

      img.src = URL.createObjectURL(blob);
    };

    ws.onclose = () => {
      setIsVideoConnected(false);
      setTimeout(connectVideoStream, 3000);
    };
  };

  const handleReturnHome = async () => {
    if (!droneId) return;
    
    try {
      await dronesAPI.sendCommand(droneId, 'return_home');
    } catch (error) {
      // Silent
    }
  };

  if (loading) {
    return <div className="loading">Загрузка...</div>;
  }

  if (!drone) {
    return <div className="error">Дрон не найден</div>;
  }

  return (
    <div className="drone-detail-page">
      <div className="page-header">
        <Button onClick={() => navigate("/drones")} variant="secondary">
          ← Назад
        </Button>
        <h1>Дрон {drone.model}</h1>
      </div>

      <div className="drone-info-card">
        <div className="info-row">
          <span className="label">ID:</span>
          <span className="value">{drone.id.substring(0, 8)}...</span>
        </div>
        <div className="info-row">
          <span className="label">Статус:</span>
          <span className={`status-badge status-${drone.status}`}>
            {drone.status}
          </span>
        </div>
        <div className="info-row">
          <span className="label">Батарея:</span>
          <span className="value">
            {drone.battery_level?.toFixed(1) || "0.0"}%
          </span>
        </div>
        <div className="info-row">
          <span className="label">IP:</span>
          <span className="value">{drone.ip_address}</span>
        </div>
      </div>

      <div className="video-section">
        <h2>Видео с камеры</h2>
        <div
          style={{
            position: "relative",
            minHeight: "400px",
            background: "#000",
            borderRadius: "8px",
          }}
        >
          <canvas
            ref={canvasRef}
            className="video-canvas"
            width={640}
            height={480}
            style={{
              width: "100%",
              height: "auto",
              background: "#000",
              display: "block",
              borderRadius: "8px",
              position: "relative",
              zIndex: 1,
            }}
          />
          {frameCount === 0 && (
            <div
              style={{
                position: "absolute",
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color: "#666",
                background: "#000",
                borderRadius: "8px",
                zIndex: 2,
                pointerEvents: "none",
              }}
            >
              Ожидание видео...
            </div>
          )}
        </div>
        <div style={{ marginTop: "8px", color: "#999", fontSize: "12px" }}>
          Статус: {isVideoConnected ? "Подключено" : "Отключено"} | Кадров
          получено: {frameCount}
        </div>
      </div>

      <div className="control-panel">
        <h2>Управление</h2>
        <Button onClick={handleReturnHome} variant="danger">
          Вернуться на базу
        </Button>
      </div>
    </div>
  );
};

export default DroneDetailPage;
