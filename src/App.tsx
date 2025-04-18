import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material';
import { Layout } from './components/Layout';
import { BaseView } from './pages/BaseView';
import { TableView } from './pages/TableView';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

const App: React.FC = () => {
  return (
    <ThemeProvider theme={theme}>
      <Router>
        <Layout>
          <Routes>
            <Route path="/bases/:baseId" element={<BaseView />} />
            <Route path="/bases/:baseId/tables/:tableId" element={<TableView />} />
            <Route path="/" element={<div>Select a base to get started</div>} />
          </Routes>
        </Layout>
      </Router>
    </ThemeProvider>
  );
};

export default App;
