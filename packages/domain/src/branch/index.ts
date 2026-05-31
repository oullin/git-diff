export interface Branch {
  name: string;
  locked: boolean;
  lockedBy?: number;
  lockedAt?: string;
  lastSeenAt: string;
}
