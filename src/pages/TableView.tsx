import React from 'react';
import { useParams } from 'react-router-dom';
import {
  Box,
  Typography,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  IconButton,
  Tooltip,
} from '@mui/material';
import AddBoxIcon from '@mui/icons-material/AddBox';
import FilterListIcon from '@mui/icons-material/FilterList';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getPaginationRowModel,
  ColumnDef,
  flexRender,
  SortingState,
  PaginationState,
} from '@tanstack/react-table';
import { Field, Record as RecordType, FieldType } from '../types';
import { fieldApi, recordApi } from '../services/api';
import { websocketService } from '../services/websocket';

export const TableView: React.FC = () => {
  const { baseId, tableId } = useParams<{ baseId: string; tableId: string }>();
  const [fields, setFields] = React.useState<Field[]>([]);
  const [records, setRecords] = React.useState<RecordType[]>([]);
  const [totalRecords, setTotalRecords] = React.useState(0);
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [pagination, setPagination] = React.useState<PaginationState>({
    pageIndex: 0,
    pageSize: 10,
  });
  const [openAddFieldDialog, setOpenAddFieldDialog] = React.useState(false);
  const [newFieldName, setNewFieldName] = React.useState('');
  const [newFieldType, setNewFieldType] = React.useState<FieldType>('text');
  const [newFieldKeyName, setNewFieldKeyName] = React.useState('');

  React.useEffect(() => {
    const fetchFields = async () => {
      if (baseId && tableId) {
        try {
          const response = await fieldApi.getAll(baseId, tableId);
          setFields(response.data);
        } catch (error) {
          console.error('Error fetching fields:', error);
        }
      }
    };

    fetchFields();
  }, [baseId, tableId]);

  const fetchRecords = async () => {
    if (baseId && tableId) {
      try {
        const response = await recordApi.getAll(baseId, tableId, {
          pagination: {
            limit: pagination.pageSize,
            offset: pagination.pageIndex * pagination.pageSize,
          },
          sort: sorting.map(s => ({
            field: s.id,
            direction: s.desc ? 'desc' : 'asc',
          })),
        });
        setRecords(response.data.records);
        setTotalRecords(response.data.total);
      } catch (error) {
        console.error('Error fetching records:', error);
      }
    }
  };

  React.useEffect(() => {
    fetchRecords();
  }, [baseId, tableId, pagination, sorting]);

  React.useEffect(() => {
    if (tableId) {
      websocketService.connect(tableId, (message) => {
        fetchRecords();
      });

      return () => {
        websocketService.disconnect();
      };
    }
  }, [tableId]);

  const handleCreateField = async () => {
    if (baseId && tableId && newFieldName && newFieldType && newFieldKeyName) {
      try {
        const response = await fieldApi.create(baseId, tableId, {
          Name: newFieldName,
          Type: newFieldType,
          KeyName: newFieldKeyName,
        });
        setFields([...fields, response.data]);
        setNewFieldName('');
        setNewFieldType('text');
        setNewFieldKeyName('');
        setOpenAddFieldDialog(false);
      } catch (error) {
        console.error('Error creating field:', error);
      }
    }
  };

  const handleCreateRecord = async () => {
    if (baseId && tableId) {
      try {
        const defaultData = fields.reduce((acc, field) => {
          acc[field.KeyName] = '';
          return acc;
        }, {} as Record<string, string>);

        await recordApi.create(baseId, tableId, defaultData);
        fetchRecords();
      } catch (error) {
        console.error('Error creating record:', error);
      }
    }
  };

  const handleUpdateRecord = async (recordId: string, data: Record<string, string>) => {
    if (baseId && tableId) {
      try {
        await recordApi.update(baseId, tableId, recordId, data);
        fetchRecords();
      } catch (error) {
        console.error('Error updating record:', error);
      }
    }
  };

  const handleDeleteRecord = async (recordId: string) => {
    if (baseId && tableId && window.confirm('Are you sure you want to delete this record?')) {
      try {
        await recordApi.delete(baseId, tableId, recordId);
        fetchRecords();
      } catch (error) {
        console.error('Error deleting record:', error);
      }
    }
  };

  const columns: ColumnDef<any>[] = [
    ...fields.map(field => ({
      accessorKey: field.KeyName,
      header: field.Name,
    })),
    {
      id: 'actions',
      header: 'Actions',
      cell: ({ row }) => (
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
          <IconButton
            size="small"
            onClick={() => handleUpdateRecord(row.original.id, row.original.record.Data)}
          >
            <EditIcon fontSize="small" />
          </IconButton>
          <IconButton
            size="small"
            onClick={() => handleDeleteRecord(row.original.id)}
          >
            <DeleteIcon fontSize="small" />
          </IconButton>
        </Box>
      ),
    },
  ];

  const tableData = records.map(record => {
    const rowData: { [key: string]: any } = {
      id: record.ID,
      record: record,
    };

    fields.forEach(field => {
      rowData[field.KeyName] = record.Data[field.KeyName] || '-';
    });

    return rowData;
  });

  const table = useReactTable({
    data: tableData,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    state: {
      sorting,
      pagination,
    },
    onSortingChange: setSorting,
    onPaginationChange: setPagination,
  });

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">Table Data</Typography>
        <Box>
          <Tooltip title="Add Column">
            <IconButton onClick={() => setOpenAddFieldDialog(true)}>
              <AddBoxIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Add Filter">
            <IconButton>
              <FilterListIcon />
            </IconButton>
          </Tooltip>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={handleCreateRecord}
            sx={{ ml: 1 }}
          >
            Add Record
          </Button>
        </Box>
      </Box>

      <Box sx={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            {table.getHeaderGroups().map(headerGroup => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map(header => (
                  <th
                    key={header.id}
                    style={{
                      padding: '8px',
                      border: '1px solid #ddd',
                      textAlign: 'left',
                    }}
                  >
                    {flexRender(
                      header.column.columnDef.header,
                      header.getContext()
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map((row) => (
              <tr key={row.id}>
                {row.getVisibleCells().map((cell) => (
                  <td
                    key={cell.id}
                    style={{
                      padding: '8px',
                      border: '1px solid #ddd',
                    }}
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </Box>

      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 2 }}>
        <Box>
          <Button
            onClick={() => table.setPageIndex(0)}
            disabled={!table.getCanPreviousPage()}
          >
            {'<<'}
          </Button>
          <Button
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            {'<'}
          </Button>
          <Button
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            {'>'}
          </Button>
          <Button
            onClick={() => table.setPageIndex(table.getPageCount() - 1)}
            disabled={!table.getCanNextPage()}
          >
            {'>>'}
          </Button>
        </Box>
        <Box>
          <span>
            Page{' '}
            <strong>
              {table.getState().pagination.pageIndex + 1} of{' '}
              {table.getPageCount()}
            </strong>
          </span>
          <select
            value={table.getState().pagination.pageSize}
            onChange={(e) => {
              table.setPageSize(Number(e.target.value));
            }}
          >
            {[10, 20, 30, 40, 50].map((pageSize) => (
              <option key={pageSize} value={pageSize}>
                Show {pageSize}
              </option>
            ))}
          </select>
        </Box>
      </Box>

      <Dialog open={openAddFieldDialog} onClose={() => setOpenAddFieldDialog(false)}>
        <DialogTitle>Add Column</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Field Name"
            fullWidth
            value={newFieldName}
            onChange={(e) => setNewFieldName(e.target.value)}
          />
          <TextField
            margin="dense"
            label="Key Name"
            fullWidth
            value={newFieldKeyName}
            onChange={(e) => setNewFieldKeyName(e.target.value)}
          />
          <FormControl fullWidth margin="dense">
            <InputLabel>Field Type</InputLabel>
            <Select
              value={newFieldType}
              onChange={(e) => setNewFieldType(e.target.value as FieldType)}
            >
              <MenuItem value="text">Text</MenuItem>
              <MenuItem value="number">Number</MenuItem>
              <MenuItem value="boolean">Boolean</MenuItem>
              <MenuItem value="date">Date</MenuItem>
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenAddFieldDialog(false)}>Cancel</Button>
          <Button onClick={handleCreateField} variant="contained">
            Create
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
