import { useState, useEffect } from "react";

interface NetworkStatus {
  isOnline: boolean;
  wasOffline: boolean; // Was offline and just recovered
}

export function useNetworkStatus(): NetworkStatus {
  const [status, setStatus] = useState<NetworkStatus>({
    isOnline: navigator.onLine,
    wasOffline: false,
  });

  useEffect(() => {
    const handleOnline = () => {
      setStatus((prev) => ({
        isOnline: true,
        wasOffline: !prev.isOnline, // True if was previously offline
      }));

      // Reset wasOffline after 3 seconds
      setTimeout(() => {
        setStatus((prev) => ({ ...prev, wasOffline: false }));
      }, 3000);
    };

    const handleOffline = () => {
      setStatus({ isOnline: false, wasOffline: false });
    };

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  return status;
}
