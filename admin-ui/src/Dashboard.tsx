import { useEffect, useState } from 'react';
import { Title } from 'react-admin';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Grid from '@mui/material/Grid';

interface Analytics {
  total_users: number;
  active_subscriptions: number;
  active_promo_codes: number;
  total_modules: number;
  total_test_attempts: number;
}

const StatCard = ({ label, value }: { label: string; value: number }) => (
  <Card sx={{ textAlign: 'center' }}>
    <CardContent>
      <Typography variant="h4" color="primary">{value}</Typography>
      <Typography variant="body2" color="text.secondary">{label}</Typography>
    </CardContent>
  </Card>
);

const Dashboard = () => {
  const [analytics, setAnalytics] = useState<Analytics | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('admin_token') || '';
    fetch('/api/v1/admin/analytics/overview', {
      headers: { 'X-Admin-Token': token },
    })
      .then((r) => r.json())
      .then(setAnalytics)
      .catch(console.error);
  }, []);

  return (
    <div>
      <Title title="Дашборд — Mnemo Admin" />
      <Grid container spacing={2} sx={{ mt: 1 }}>
        {analytics && (
          <>
            <Grid size={{ xs: 12, sm: 6, md: 4, lg: 2 }}>
              <StatCard label="Пользователи" value={analytics.total_users} />
            </Grid>
            <Grid size={{ xs: 12, sm: 6, md: 4, lg: 2 }}>
              <StatCard label="Активные подписки" value={analytics.active_subscriptions} />
            </Grid>
            <Grid size={{ xs: 12, sm: 6, md: 4, lg: 2 }}>
              <StatCard label="Активные промокоды" value={analytics.active_promo_codes} />
            </Grid>
            <Grid size={{ xs: 12, sm: 6, md: 4, lg: 2 }}>
              <StatCard label="Модули" value={analytics.total_modules} />
            </Grid>
            <Grid size={{ xs: 12, sm: 6, md: 4, lg: 2 }}>
              <StatCard label="Попытки тестов" value={analytics.total_test_attempts} />
            </Grid>
          </>
        )}
      </Grid>
    </div>
  );
};

export default Dashboard;
