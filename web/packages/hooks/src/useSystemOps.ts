import { api } from '@zippyra/api-client';

export interface DLQTopicSummary {
  topic_name: string;
  message_count: number;
  oldest_message_age_seconds: number;
}

export interface DLQMessage {
  topic: string;
  offset: number;
  key: string;
  value: any;
  headers?: Record<string, string>;
  timestamp: string;
}

export interface FeatureFlag {
  flag_key: string;
  description: string;
  scope_type: 'GLOBAL' | 'CHAIN' | 'STORE' | 'USER_PERCENTAGE';
  enabled_globally: boolean;
  enabled_scope_ids?: string[];
  user_percentage?: number;
  updated_by: string;
  updated_at: string;
  created_at: string;
}

export interface CircuitBreakerStatus {
  gateway: string;
  state: 'CLOSED' | 'OPEN' | 'HALF_OPEN';
  error_rate_rolling_1min: number;
  opened_at?: string;
  will_retry_at?: string;
}

export function useSystemOps() {
  // Kafka DLQ
  const getDlqTopics = async () => {
    return api.get<{ dlq_topics: DLQTopicSummary[] }>('/v1/audit/kafka/dlq-topics');
  };

  const peekDlqMessages = async (topic: string, limit: number = 20) => {
    return api.get<{ topic: string; messages: DLQMessage[]; total: number }>(
      `/v1/audit/kafka/dlq-topics/${encodeURIComponent(topic)}/messages?limit=${limit}`
    );
  };

  const replayDlqMessages = async (topic: string, offsets: number[]) => {
    return api.post<{ replayed_count: number; failed_offsets: number[] }>(
      `/v1/audit/kafka/dlq-topics/${encodeURIComponent(topic)}/replay`,
      { offsets }
    );
  };

  const discardDlqMessages = async (topic: string, offsets: number[], reason: string, stepUpToken?: string) => {
    const headers: Record<string, string> = {};
    if (stepUpToken) {
      headers['X-StepUp-Token'] = stepUpToken;
    }
    return api.delete<{ status: string; discarded_offsets: number[] }>(
      `/v1/audit/kafka/dlq-topics/${encodeURIComponent(topic)}/messages`,
      { offsets, reason },
      { headers }
    );
  };

  // Feature Flags
  const getFeatureFlags = async () => {
    return api.get<{ feature_flags: FeatureFlag[] }>('/v1/audit/feature-flags');
  };

  const createFeatureFlag = async (payload: { flag_key: string; description: string; scope_type: string }) => {
    return api.post<FeatureFlag>('/v1/audit/feature-flags', payload);
  };

  const updateFeatureFlag = async (key: string, payload: Partial<FeatureFlag>, stepUpToken?: string) => {
    const headers: Record<string, string> = {};
    if (stepUpToken) {
      headers['X-StepUp-Token'] = stepUpToken;
    }
    return api.put<FeatureFlag>(`/v1/audit/feature-flags/${encodeURIComponent(key)}`, payload, { headers });
  };

  const deleteFeatureFlag = async (key: string) => {
    return api.delete<{ status: string; flag_key: string }>(`/v1/audit/feature-flags/${encodeURIComponent(key)}`);
  };

  // Circuit Breaker Status
  const getCircuitBreakerStatus = async () => {
    return api.get<CircuitBreakerStatus>('/v1/payment/internal/circuit-breaker-status');
  };

  return {
    getDlqTopics,
    peekDlqMessages,
    replayDlqMessages,
    discardDlqMessages,
    getFeatureFlags,
    createFeatureFlag,
    updateFeatureFlag,
    deleteFeatureFlag,
    getCircuitBreakerStatus,
  };
}
