import { List, Datagrid, TextField, DateField, ShowButton, Show, SimpleShowLayout } from 'react-admin';

export const FeedbackList = () => (
  <List sort={{ field: 'created_at', order: 'DESC' }} exporter={false}>
    <Datagrid bulkActionButtons={false}>
      <TextField source="id" label="ID" />
      <TextField source="user_id" label="Telegram ID" />
      <TextField source="text" label="Сообщение" />
      <DateField source="created_at" label="Дата" showTime />
      <ShowButton />
    </Datagrid>
  </List>
);

export const FeedbackShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID" />
      <TextField source="user_id" label="Telegram ID" />
      <TextField source="text" label="Сообщение" />
      <DateField source="created_at" label="Дата" showTime />
    </SimpleShowLayout>
  </Show>
);
