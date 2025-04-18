import React from 'react';
import { useParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { useToast } from '@/components/ui/use-toast';

import { RecordList } from '../components/RecordList';
import { RecordDialog } from '../components/RecordDialog';
import { fieldService } from '../services/field';
import { Field } from '../types/field';
import { Record } from '../types/record';

export default function TablePage() {
  const params = useParams();
  const tableId = params.tableId as string;
  const { toast } = useToast();
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [selectedRecord, setSelectedRecord] = React.useState<Record | undefined>();

  const { data: fields, isLoading: isLoadingFields } = useQuery<Field[]>({
    queryKey: ['fields', tableId],
    queryFn: () => fieldService.getFieldsByTableId(tableId),
  });

  if (isLoadingFields) {
    return <div>Loading...</div>;
  }

  if (!fields) {
    return <div>No fields found</div>;
  }

  const handleCreateRecord = () => {
    setSelectedRecord(undefined);
    setDialogOpen(true);
  };

  const handleEditRecord = (record: Record) => {
    setSelectedRecord(record);
    setDialogOpen(true);
  };

  const handleDialogSuccess = () => {
    // The RecordList component will automatically update via WebSocket
  };

  return (
    <div className="container mx-auto py-6">
      <RecordList
        tableId={tableId}
        fields={fields}
        onCreateRecord={handleCreateRecord}
        onEditRecord={handleEditRecord}
      />
      <RecordDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        tableId={tableId}
        fields={fields}
        record={selectedRecord}
        onSuccess={handleDialogSuccess}
      />
    </div>
  );
} 