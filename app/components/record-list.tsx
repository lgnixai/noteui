'use client';

import React, { useState } from 'react';
import { Record, Field } from '@/types/record';
import { Button } from '@/components/ui/button';
import { Plus, Pencil, Trash2 } from 'lucide-react';
import { RecordDialog } from './record-dialog';
import { useToast } from '@/components/ui/use-toast';
import { recordService } from '@/services/record';
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  ColumnDef,
} from '@tanstack/react-table';

interface RecordListProps {
  tableId: string;
  fields: Field[];
  records: Record[];
  total: number;
  onRefresh: () => void;
  onEdit: (record: Record) => void;
  onDelete: (recordId: string) => void;
}

export function RecordList({ tableId, fields, records, total, onRefresh, onEdit, onDelete }: RecordListProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Record | null>(null);
  const { toast } = useToast();

  // Transform records into table data format
  const tableData = records.map(record => {
    const rowData: { [key: string]: any } = {
      id: record.ID,
      record: record, // Keep the original record for actions
    };
    
    // Add field values
    fields.forEach(field => {
      rowData[field.KeyName] = record.Data[field.KeyName] || '-';
    });
    
    return rowData;
  });

  // Create table columns
  const columns: ColumnDef<any>[] = [
    ...fields.map(field => ({
      accessorKey: field.KeyName,
      header: field.Name,
    })),
    {
      id: 'actions',
      header: 'Actions',
      cell: ({ row }) => (
        <div className="flex justify-end space-x-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onEdit(row.original.record)}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete(row.original.id)}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ),
    },
  ];

  const table = useReactTable({
    data: tableData,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  const handleCreate = () => {
    setEditingRecord(null);
    setIsDialogOpen(true);
  };

  const handleSubmit = async (data: Record['Data']) => {
    try {
      if (editingRecord) {
        await recordService.updateRecord(editingRecord.ID, { data });
        toast({
          title: 'Success',
          description: 'Record updated successfully',
        });
      } else {
        await recordService.createRecord(tableId, { data });
        toast({
          title: 'Success',
          description: 'Record created successfully',
        });
      }
      setIsDialogOpen(false);
      onRefresh();
    } catch (error) {
      toast({
        title: 'Error',
        description: editingRecord ? 'Failed to update record' : 'Failed to create record',
        variant: 'destructive',
      });
    }
  };

  if (!records.length) {
    return (
      <div className="text-center py-8 text-gray-500">
        No records found
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Records ({total})</h2>
        <Button onClick={handleCreate}>
          <Plus className="mr-2 h-4 w-4" />
          Add Record
        </Button>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            {table.getHeaderGroups().map(headerGroup => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map(header => (
                  <th
                    key={header.id}
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
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
          <tbody className="bg-white divide-y divide-gray-200">
            {table.getRowModel().rows.map(row => (
              <tr key={row.id}>
                {row.getVisibleCells().map(cell => (
                  <td
                    key={cell.id}
                    className="px-6 py-4 whitespace-nowrap text-sm text-gray-900"
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <RecordDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        fields={fields}
        record={editingRecord}
        onSubmit={handleSubmit}
      />
    </div>
  );
} 