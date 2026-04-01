import {
  List, Datagrid, TextField, DateField, FunctionField,
  Create, Edit, SimpleForm, NumberInput, SelectInput,
  required, EditButton,
} from 'react-admin';

const roleChoices = [
  { id: 'student', name: 'Студент' },
  { id: 'teacher', name: 'Преподаватель' },
];

const subStatusChoices = [
  { id: 'inactive', name: 'Неактивна' },
  { id: 'active', name: 'Активна' },
  { id: 'expired', name: 'Истекла' },
];

export const UserList = () => (
  <List sort={{ field: 'telegram_id', order: 'ASC' }}>
    <Datagrid bulkActionButtons={false}>
      <TextField source="telegram_id" label="Telegram ID" />
      <FunctionField
        label="Username"
        render={(record: any) => record.username ? `@${record.username}` : '—'}
      />
      <TextField source="role" label="Роль" />
      <TextField source="subscription_status" label="Подписка" />
      <DateField source="created_at" label="Зарегистрирован" />
      <EditButton />
    </Datagrid>
  </List>
);

export const UserCreate = () => (
  <Create redirect="list">
    <SimpleForm>
      <NumberInput source="telegram_id" label="Telegram ID" validate={required()} />
      <SelectInput source="role" label="Роль" choices={roleChoices} defaultValue="student" />
      <SelectInput source="subscription_status" label="Подписка" choices={subStatusChoices} defaultValue="inactive" />
    </SimpleForm>
  </Create>
);

export const UserEdit = () => (
  <Edit>
    <SimpleForm>
      <SelectInput source="role" label="Роль" choices={roleChoices} />
      <SelectInput source="subscription_status" label="Подписка" choices={subStatusChoices} />
    </SimpleForm>
  </Edit>
);
