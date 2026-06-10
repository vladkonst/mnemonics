import {
  List,
  Datagrid,
  TextField,
  NumberField,
  DateField,
  ShowButton,
  Show,
  SimpleShowLayout,
} from 'react-admin';

export const CorporateGroupList = () => (
  <List sort={{ field: 'created_at', order: 'DESC' }} exporter={false}>
    <Datagrid bulkActionButtons={false}>
      <TextField source="id" label="ID группы" />
      <TextField source="name" label="Название" />
      <TextField source="purchase_id" label="ID покупки" />
      <TextField source="manager_id" label="Менеджер (TG ID)" />
      <TextField source="teacher_id" label="Преподаватель (TG ID)" />
      <NumberField source="semesters" label="Семестров" />
      <DateField source="created_at" label="Создана" showTime />
      <ShowButton />
    </Datagrid>
  </List>
);

export const CorporateGroupShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID группы" />
      <TextField source="name" label="Название" />
      <TextField source="purchase_id" label="ID покупки" />
      <TextField source="manager_id" label="Менеджер (Telegram ID)" />
      <TextField source="teacher_id" label="Преподаватель (Telegram ID)" />
      <TextField source="teacher_join_code" label="Код преподавателя (join code)" />
      <TextField source="student_link_id" label="ID студенческой ссылки" />
      <NumberField source="semesters" label="Количество семестров" />
      <DateField source="created_at" label="Создана" showTime />
    </SimpleShowLayout>
  </Show>
);
