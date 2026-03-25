import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type { Notification } from '@/types/api';

/** Fetch recent notifications for the authenticated user. */
export function useNotifications(limit = 20) {
  return useQuery({
    queryKey: queryKeys.notifications.list(),
    queryFn: () => apiClient.get<Notification[]>('/api/notifications', { limit: String(limit) }),
    staleTime: staleTimes.notifications,
  });
}

/** Fetch unread notification count. */
export function useUnreadCount() {
  return useQuery({
    queryKey: queryKeys.notifications.unreadCount(),
    queryFn: () => apiClient.get<{ count: number }>('/api/notifications/unread-count'),
    staleTime: staleTimes.notifications,
    refetchInterval: 30_000, // Poll every 30 seconds
  });
}

/** Mark a notification as read. */
export function useMarkNotificationRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiClient.put<{ message: string }>(`/api/notifications/${id}/read`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.notifications.all });
    },
  });
}
