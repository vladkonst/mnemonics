import {
  List,
  Datagrid,
  TextField,
  DateField,
  ShowButton,
  Show,
  SimpleShowLayout,
  ArrayField,
  NumberField,
} from 'react-admin';

export const InviteLinkList = () => (
  <List sort={{ field: 'created_at', order: 'DESC' }} exporter={false}>
    <Datagrid bulkActionButtons={false}>
      <TextField source="id" label="ID ссылки" />
      <TextField source="teacher_id" label="Telegram преподавателя" />
      <TextField source="promo_code" label="Промокод" />
      <DateField source="created_at" label="Создана" showTime />
      <ShowButton />
    </Datagrid>
  </List>
);

export const InviteLinkShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID ссылки" />
      <TextField source="teacher_id" label="Telegram преподавателя" />
      <TextField source="promo_code" label="Промокод" />
      <NumberField source="activation_count" label="Всего активаций" />
      <DateField source="created_at" label="Создана" showTime />
      <ArrayField source="activations">
        <Datagrid bulkActionButtons={false}>
          <TextField source="user_id" label="Telegram ID" />
          <DateField source="activated_at" label="Дата активации" showTime />
        </Datagrid>
      </ArrayField>
    </SimpleShowLayout>
  </Show>
);
