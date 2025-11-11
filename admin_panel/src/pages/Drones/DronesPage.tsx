import { useEffect, useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { dronesAPI } from "../../api";
import type { Drone, CreateDroneRequest } from "../../types";
import Button from "../../components/Button/Button";
import Modal from "../../components/Modal/Modal";
import ConfirmModal from "../../components/ConfirmModal/ConfirmModal";
import { showSuccess, showError } from "../../utils/toast";
import "./DronesPage.css";

const DronesPage = () => {
  const navigate = useNavigate();
  const [drones, setDrones] = useState<Drone[]>([]);
  const [loading, setLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false);
  const [droneToDelete, setDroneToDelete] = useState<Drone | null>(null);
  const [editingDrone, setEditingDrone] = useState<Drone | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [formData, setFormData] = useState<
    CreateDroneRequest & { status: string }
  >({
    model: "",
    ip_address: "",
    status: "idle",
  });
  const [editFormData, setEditFormData] = useState({
    model: "",
    ip_address: "",
  });

  useEffect(() => {
    loadDrones();
    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const connectWebSocket = () => {
    const ws = new WebSocket("ws://localhost:8081/ws/admin");
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("Connected to drone-service WebSocket");
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === "drones_status" && data.drones) {
          setDrones((prevDrones) =>
            prevDrones.map((drone) => {
              const update = data.drones.find(
                (d: any) => d.drone_id === drone.id
              );
              if (update) {
                return {
                  ...drone,
                  battery_level: update.battery_level,
                  status: update.status,
                  current_delivery_id: update.current_delivery_id,
                  speed: update.speed,
                  error_message: update.error_message,
                };
              }
              return drone;
            })
          );
        }
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected, reconnecting in 3s...");
      setTimeout(connectWebSocket, 3000);
    };
  };

  const loadDrones = async () => {
    try {
      setLoading(true);
      const response = await dronesAPI.getAll();
      setDrones(response.data);
    } catch (error) {
      console.error("Failed to load drones:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      await dronesAPI.create({
        model: formData.model,
        ip_address: formData.ip_address,
      });
      showSuccess("Дрон успешно создан!");
      await loadDrones();
      closeModal();
    } catch (error: any) {
      console.error("Failed to save drone:", error);
      const errorMsg =
        error.response?.data?.error || error.message || "Неизвестная ошибка";
      showError(`Ошибка: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (drone: Drone) => {
    setDroneToDelete(drone);
    setIsConfirmModalOpen(true);
  };

  const confirmDelete = async () => {
    if (!droneToDelete) return;
    setLoading(true);
    try {
      await dronesAPI.delete(droneToDelete.id);
      showSuccess("Дрон успешно удален!");
      await loadDrones();
    } catch (error: any) {
      console.error("Failed to delete drone:", error);
      const errorMsg =
        error.response?.data?.error || error.message || "Неизвестная ошибка";
      showError(`Ошибка при удалении: ${errorMsg}`);
    } finally {
      setLoading(false);
      setIsConfirmModalOpen(false);
      setDroneToDelete(null);
    }
  };

  const openModal = () => {
    setFormData({ model: "", ip_address: "", status: "idle" });
    setIsModalOpen(true);
  };

  const openEditModal = (drone: Drone) => {
    setEditingDrone(drone);
    setEditFormData({ model: drone.model, ip_address: drone.ip_address || "" });
    setIsEditModalOpen(true);
  };

  const closeEditModal = () => {
    setIsEditModalOpen(false);
    setEditingDrone(null);
    setEditFormData({ model: "", ip_address: "" });
  };

  const handleEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingDrone) return;
    setLoading(true);

    try {
      await dronesAPI.update(editingDrone.id, editFormData);
      showSuccess("Дрон успешно обновлен!");
      await loadDrones();
      closeEditModal();
    } catch (error: any) {
      console.error("Failed to update drone:", error);
      const errorMsg =
        error.response?.data?.error || error.message || "Неизвестная ошибка";
      showError(`Ошибка: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setFormData({ model: "", ip_address: "", status: "idle" });
  };

  const getStatusBadge = (status: string) => {
    const badges: Record<string, { label: string; className: string }> = {
      idle: { label: "Свободен", className: "status-idle" },
      busy: { label: "Занят", className: "status-busy" },
      delivering: { label: "Доставляет", className: "status-busy" },
      error: { label: "Ошибка", className: "status-error" },
    };
    const badge = badges[status] || {
      label: "Неизвестно",
      className: "status-error",
    };
    return (
      <span className={`status-badge ${badge.className}`}>{badge.label}</span>
    );
  };

  return (
    <div className="drones-page">
      <div className="page-header">
        <div>
          <h1>🚁 Управление Дронами</h1>
          <p>Просмотр и редактирование дронов</p>
        </div>
        <Button onClick={() => openModal()}>➕ Добавить дрон</Button>
      </div>

      {loading ? (
        <div className="loading">Загрузка...</div>
      ) : (
        <div className="drones-grid">
          {drones.map((drone) => (
            <div
              key={drone.id}
              className="drone-card"
              onClick={() => navigate(`/drones/${drone.id}`)}
            >
              <div className="drone-header">
                <h3>{drone.model}</h3>
              </div>
              <div className="drone-id">ID: {drone.id.substring(0, 8)}...</div>
              <div className="drone-ip">IP: {drone.ip_address}</div>
              <div className="drone-battery">
                🔋 Батарея:{" "}
                <strong>{drone.battery_level?.toFixed(1) || "0.0"}%</strong>
              </div>
              <div className="drone-status-display">
                Статус: {getStatusBadge(drone.status)}
              </div>
              <div
                className="drone-actions"
                onClick={(e) => e.stopPropagation()}
              >
                <Button
                  size="small"
                  variant="primary"
                  onClick={() => openEditModal(drone)}
                  disabled={loading}
                >
                  ✏️ Изменить
                </Button>
                <Button
                  size="small"
                  variant="danger"
                  onClick={() => handleDelete(drone)}
                  disabled={loading}
                >
                  🗑️ Удалить
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {drones.length === 0 && !loading && (
        <div className="empty-state">
          <p>Нет дронов. Добавьте первый дрон!</p>
        </div>
      )}

      <Modal isOpen={isModalOpen} onClose={closeModal} title="Добавить дрон">
        <form onSubmit={handleSubmit} className="drone-form">
          <div className="form-group">
            <label>Модель</label>
            <input
              type="text"
              value={formData.model}
              onChange={(e) =>
                setFormData({ ...formData, model: e.target.value })
              }
              required
              placeholder="Например: DJI Phantom 5"
            />
          </div>

          <div className="form-group">
            <label>IP-адрес</label>
            <input
              type="text"
              value={formData.ip_address}
              onChange={(e) =>
                setFormData({ ...formData, ip_address: e.target.value })
              }
              placeholder="Например: 192.168.10.1"
            />
          </div>

          <div className="form-actions">
            <Button
              type="button"
              variant="secondary"
              onClick={closeModal}
              disabled={loading}
            >
              Отмена
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? "⏳ Загрузка..." : "Создать"}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isEditModalOpen}
        onClose={closeEditModal}
        title="Редактировать дрон"
      >
        <form onSubmit={handleEdit} className="drone-form">
          <div className="form-group">
            <label>Модель</label>
            <input
              type="text"
              value={editFormData.model}
              onChange={(e) =>
                setEditFormData({ ...editFormData, model: e.target.value })
              }
              required
              placeholder="Например: DJI Phantom 5"
            />
          </div>

          <div className="form-group">
            <label>IP-адрес</label>
            <input
              type="text"
              value={editFormData.ip_address}
              onChange={(e) =>
                setEditFormData({ ...editFormData, ip_address: e.target.value })
              }
              placeholder="Например: 192.168.10.1"
            />
          </div>

          <div className="form-actions">
            <Button
              type="button"
              variant="secondary"
              onClick={closeEditModal}
              disabled={loading}
            >
              Отмена
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? "⏳ Загрузка..." : "Сохранить"}
            </Button>
          </div>
        </form>
      </Modal>

      <ConfirmModal
        isOpen={isConfirmModalOpen}
        onClose={() => {
          setIsConfirmModalOpen(false);
          setDroneToDelete(null);
        }}
        onConfirm={confirmDelete}
        title="Подтверждение удаления"
        message={
          droneToDelete ? (
            <div>
              <p>Вы уверены, что хотите удалить дрон?</p>
              <p>
                <strong>{droneToDelete.model}</strong>
              </p>
              <p>
                ID: <strong>{droneToDelete.id.substring(0, 8)}...</strong>
              </p>
            </div>
          ) : (
            ""
          )
        }
        confirmText="Удалить"
        cancelText="Отмена"
        variant="danger"
        loading={loading}
      />
    </div>
  );
};

export default DronesPage;
