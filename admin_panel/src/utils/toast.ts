import toast from 'react-hot-toast';

export const showSuccess = (message: string) => {
  toast.success(message, {
    duration: 4000,
    position: 'top-center',
    style: {
      background: '#0f0e17',
      color: '#fffffe',
      border: '1px solid #00d4aa',
      marginTop: '80px',
    },
    iconTheme: {
      primary: '#00d4aa',
      secondary: '#0f0e17',
    },
  });
};

export const showError = (message: string) => {
  toast.error(message, {
    duration: 5000,
    position: 'top-center',
    style: {
      background: '#0f0e17',
      color: '#fffffe',
      border: '1px solid #ff6b6b',
      marginTop: '80px',
    },
    iconTheme: {
      primary: '#ff6b6b',
      secondary: '#0f0e17',
    },
  });
};

export const showInfo = (message: string) => {
  toast(message, {
    duration: 4000,
    position: 'top-center',
    style: {
      background: '#0f0e17',
      color: '#fffffe',
      border: '1px solid #6c5ce7',
      marginTop: '80px',
    },
    icon: 'ℹ️',
  });
};

