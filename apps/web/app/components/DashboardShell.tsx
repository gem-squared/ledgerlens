'use client';

import { useState } from 'react';
import { JudgeRequestConsole } from './JudgeRequestConsole';
import { StatsDashboard } from './StatsDashboard';
import { RecentActivity } from './RecentActivity';
import { JudgeAuditPanel } from './JudgeAuditPanel';

// Client-side shell that wires the post-run "counter ticks +1" loop:
//   onRunComplete in JudgeRequestConsole increments refreshTrigger,
//   which StatsDashboard + RecentActivity observe and refetch.
export function DashboardShell() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const bumpStats = () => setRefreshTrigger((n) => n + 1);

  return (
    <div className="space-y-6">
      <StatsDashboard refreshTrigger={refreshTrigger} />

      <JudgeRequestConsole onRunComplete={bumpStats} />

      <div className="grid gap-6 lg:grid-cols-2">
        <RecentActivity refreshTrigger={refreshTrigger} />
        <JudgeAuditPanel />
      </div>
    </div>
  );
}
