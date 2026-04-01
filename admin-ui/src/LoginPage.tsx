import { useState } from 'react';
import { useLogin, useNotify } from 'react-admin';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';

const LoginPage = () => {
  const [token, setToken] = useState('');
  const login = useLogin();
  const notify = useNotify();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // authProvider.login receives { username, password } — we pass token as username
    login({ username: token, password: '' }).catch(() =>
      notify('Неверный токен', { type: 'error' })
    );
  };

  return (
    <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh" bgcolor="#f5f5f5">
      <Paper sx={{ p: 4, width: 360 }}>
        <Typography variant="h5" mb={3} textAlign="center">Mnemo Admin</Typography>
        <form onSubmit={handleSubmit}>
          <TextField
            fullWidth
            label="Токен"
            type="password"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            margin="normal"
            required
          />
          <Button fullWidth variant="contained" type="submit" sx={{ mt: 2 }}>
            Войти
          </Button>
        </form>
      </Paper>
    </Box>
  );
};

export default LoginPage;
