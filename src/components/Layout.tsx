import React from 'react';
import { Box, Drawer, AppBar, Toolbar, Typography, List, ListItem, ListItemButton, ListItemText, IconButton } from '@mui/material';
import { Add as AddIcon, Edit as EditIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { Base } from '../types';
import { baseApi } from '../services/api';

const drawerWidth = 240;

export const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [bases, setBases] = React.useState<Base[]>([]);
  const navigate = useNavigate();

  React.useEffect(() => {
    const fetchBases = async () => {
      try {
        const response = await baseApi.getAll();
        setBases(response.data);
      } catch (error) {
        console.error('Error fetching bases:', error);
      }
    };

    fetchBases();
  }, []);

  const handleCreateBase = async () => {
    const name = prompt('Enter base name:');
    if (name) {
      try {
        const response = await baseApi.create({ name });
        setBases([...bases, response.data]);
      } catch (error) {
        console.error('Error creating base:', error);
      }
    }
  };

  const handleEditBase = async (base: Base) => {
    const newName = prompt('Enter new base name:', base.Name);
    if (newName && newName !== base.Name) {
      try {
        const response = await baseApi.update(base.ID, { name: newName });
        setBases(bases.map(b => b.ID === base.ID ? response.data : b));
      } catch (error) {
        console.error('Error updating base:', error);
      }
    }
  };

  const handleDeleteBase = async (base: Base) => {
    if (window.confirm(`Are you sure you want to delete "${base.Name}"?`)) {
      try {
        await baseApi.delete(base.ID);
        setBases(bases.filter(b => b.ID !== base.ID));
      } catch (error) {
        console.error('Error deleting base:', error);
      }
    }
  };

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <Typography variant="h6" noWrap component="div">
            Airtable-like App
          </Typography>
        </Toolbar>
      </AppBar>
      <Drawer
        variant="permanent"
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: drawerWidth,
            boxSizing: 'border-box',
          },
        }}
      >
        <Toolbar />
        <Box sx={{ overflow: 'auto' }}>
          <List>
            <ListItem>
              <ListItemButton onClick={handleCreateBase}>
                <AddIcon />
                <ListItemText primary="Create Base" />
              </ListItemButton>
            </ListItem>
            {bases.map((base) => (
              <ListItem
                key={base.ID}
                secondaryAction={
                  <>
                    <IconButton edge="end" onClick={() => handleEditBase(base)}>
                      <EditIcon />
                    </IconButton>
                    <IconButton edge="end" onClick={() => handleDeleteBase(base)}>
                      <DeleteIcon />
                    </IconButton>
                  </>
                }
              >
                <ListItemButton onClick={() => navigate(`/bases/${base.ID}`)}>
                  <ListItemText primary={base.Name} />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        </Box>
      </Drawer>
      <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
        <Toolbar />
        {children}
      </Box>
    </Box>
  );
};
