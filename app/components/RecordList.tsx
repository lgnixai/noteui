import React, { useEffect, useState, useCallback } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ChevronDown, ChevronUp, MoreHorizontal, Plus } from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';

import { Record, RecordSort, RecordFilter } from '../types/record';
import { Field } from '../types/field';
import { recordService } from '../services/record';
import { webSocketService } from '../services/websocket';
import { formatFieldValue } from '../utils/format';

interface RecordListProps {
  tableId: string;
  fields: Field[];
  onCreateRecord: () => void;
  onEditRecord: (record: Record) => void;
}

export function RecordList({ tableId, fields, onCreateRecord, onEditRecord }: RecordListProps) {
  const [records, setRecords] = useState<Record[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [sort, setSort] = useState<RecordSort[]>([]);
  const [filter, setFilter] = useState<RecordFilter | undefined>();
  const [page, setPage] = useState(1);
  const [searchText, setSearchText] = useState('');
  const { toast } = useToast();
  const pageSize = 50;

  const loadRecords = useCallback(async () => {
    try {
      setLoading(true);
      const offset = (page - 1) * pageSize;
      const response = await recordService.getRecords(tableId, filter, sort, pageSize, offset);
      setRecords(response.records);
      setTotal(response.total);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load records',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [tableId, filter, sort, page, toast]);

  useEffect(() => {
    loadRecords();
  }, [loadRecords]);

  useEffect(() => {
    // Connect to WebSocket and subscribe to updates
    webSocketService.connect();
    
    const unsubscribe = webSocketService.subscribe((message) => {
      if (message.tableId === tableId) {
        switch (message.type) {
          case 'record_created':
          case 'record_updated':
            if (message.record) {
              setRecords(prevRecords => {
                const index = prevRecords.findIndex(r => r.id === message.recordId);
                if (index >= 0) {
                  // Update existing record
                  const newRecords = [...prevRecords];
                  newRecords[index] = message.record;
                  return newRecords;
                } else if (message.type === 'record_created') {
                  // Add new record if we're on the first page
                  return page === 1 ? [message.record, ...prevRecords] : prevRecords;
                }
                return prevRecords;
              });
            }
            break;
          case 'record_deleted':
            setRecords(prevRecords => 
              prevRecords.filter(r => r.id !== message.recordId)
            );
            setTotal(prev => prev - 1);
            break;
        }
      }
    });

    return () => {
      unsubscribe();
    };
  }, [tableId, page]);

  const handleSort = (fieldId: string) => {
    const existingSort = sort.find(s => s.fieldId === fieldId);
    let newSort: RecordSort[];

    if (!existingSort) {
      newSort = [...sort, { fieldId, direction: 'asc' }];
    } else if (existingSort.direction === 'asc') {
      newSort = sort.map(s => 
        s.fieldId === fieldId ? { ...s, direction: 'desc' } : s
      );
    } else {
      newSort = sort.filter(s => s.fieldId !== fieldId);
    }

    setSort(newSort);
  };

  const handleSearch = () => {
    if (!searchText) {
      setFilter(undefined);
      return;
    }

    // Create a filter that searches across all text fields
    const textFields = fields.filter(f => f.type === 'text');
    if (textFields.length === 0) return;

    const conditions = textFields.map(field => ({
      fieldId: field.id,
      operator: 'contains' as const,
      value: searchText
    }));

    setFilter({
      operator: 'OR',
      conditions
    });
  };

  const handleDeleteRecord = async (recordId: string) => {
    try {
      await recordService.deleteRecord(recordId);
      toast({
        title: 'Success',
        description: 'Record deleted successfully',
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete record',
        variant: 'destructive',
      });
    }
  };

  const getSortIcon = (fieldId: string) => {
    const fieldSort = sort.find(s => s.fieldId === fieldId);
    if (!fieldSort) return null;
    return fieldSort.direction === 'asc' ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />;
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <div className="flex gap-2">
          <Input
            placeholder="Search records..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          />
          <Button onClick={handleSearch}>Search</Button>
        </div>
        <Button onClick={onCreateRecord}>
          <Plus className="w-4 h-4 mr-2" />
          Add Record
        </Button>
      </div>

      <div className="border rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              {fields.map(field => (
                <TableHead
                  key={field.id}
                  className="cursor-pointer"
                  onClick={() => handleSort(field.id)}
                >
                  <div className="flex items-center gap-2">
                    {field.name}
                    {getSortIcon(field.id)}
                  </div>
                </TableHead>
              ))}
              <TableHead className="w-[100px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={fields.length + 1} className="text-center">
                  Loading...
                </TableCell>
              </TableRow>
            ) : records.length === 0 ? (
              <TableRow>
                <TableCell colSpan={fields.length + 1} className="text-center">
                  No records found
                </TableCell>
              </TableRow>
            ) : (
              records.map(record => (
                <TableRow key={record.id}>
                  {fields.map(field => (
                    <TableCell key={field.id}>
                      {formatFieldValue(record.data[field.keyName], field.type)}
                    </TableCell>
                  ))}
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => onEditRecord(record)}>
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          className="text-red-600"
                          onClick={() => handleDeleteRecord(record.id)}
                        >
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex justify-between items-center">
        <div className="text-sm text-gray-500">
          {total} records total
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            disabled={page === 1}
            onClick={() => setPage(p => p - 1)}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            disabled={page * pageSize >= total}
            onClick={() => setPage(p => p + 1)}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
} 