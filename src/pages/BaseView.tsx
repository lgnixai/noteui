import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Typography,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  IconButton,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
} from '@mui/material';
import { Add as AddIcon, Edit as EditIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { Table } from '../types';
import { tableApi } from '../services/api';

export const BaseView: React.FC = () => {
  const { baseId } = useParams<{ baseId: string }>();
  const navigate = useNavigate();
  const [tables, setTables] = React.useState<Table[]>([]);
  const [openDialog, setOpenDialog] = React.useState(false);
  const [newTableName, setNewTableName] = React.useState('');
  const [editingTable, setEditingTable] = React.useState<Table | null>(null);

  React.useEffect(() => {
    const fetchTables = async () => {
      if (baseId) {
        try {
          const response = await tableApi.getAll(baseId);
          setTables(response.data);
        } catch (error) {
          console.error('Error fetching tables:', error);
        }
      }
    };

    fetchTables();
  }, [baseId]);

  const handleCreateTable = async () => {
    if (baseId && newTableName) {
      try {
        const response = await tableApi.create(baseId, { name: newTableName });
        setTables([...tables, response.data]);
        setNewTableName('');
        setOpenDialog(false);
      } catch (error) {
        console.error('Error creating table:', error);
      }
    }
  };

  const handleEditTable = async (table: Table) => {
    if (baseId && newTableName && newTableName !== table.name) {
      try {
        const response = await tableApi.update(baseId, table.id, { name: newTableName });
        setTables(tables.map(t => t.id === table.id ? response.data : t));
        setNewTableName('');
        setEditingTable(null);
        setOpenDialog(false);
      } catch (error) {
        console.error('Error updating table:', error);
      }
    }
  };

  const handleDeleteTable = async (table: Table) => {
    if (baseId && window.confirm(`Are you sure you want to delete "${table.Name}"?`)) {
      try {
        await tableApi.delete(baseId, table.ID);
        setTables(tables.filter(t => t.ID !== table.ID));
      } catch (error) {
        console.error('Error deleting table:', error);
      }
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">Tables</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => {
            setEditingTable(null);
            setNewTableName('');
            setOpenDialog(true);
          }}
        >
          Create Table
        </Button>
      </Box>

      <List>
        {tables.map((table) => (
          <ListItem
            key={table.ID}
            secondaryAction={
              <>
                <IconButton
                  edge="end"
                  onClick={() => {
                    setEditingTable(table);
                    setNewTableName(table.Name);
                    setOpenDialog(true);
                  }}
                >
                  <EditIcon />
                </IconButton>
                <IconButton edge="end" onClick={() => handleDeleteTable(table)}>
                  <DeleteIcon />
                </IconButton>
              </>
            }
          >
            <ListItemButton onClick={() => navigate(`/bases/${baseId}/tables/${table.ID}`)}>
              <ListItemText primary={table.Name} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>

      <Dialog open={openDialog} onClose={() => setOpenDialog(false)}>
        <DialogTitle>{editingTable ? 'Edit Table' : 'Create Table'}</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Table Name"
            fullWidth
            value={newTableName}
            onChange={(e) => setNewTableName(e.target.value)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button
            onClick={() => {
              if (editingTable) {
                handleEditTable(editingTable);
              } else {
                handleCreateTable();
              }
            }}
            variant="contained"
          >
            {editingTable ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
