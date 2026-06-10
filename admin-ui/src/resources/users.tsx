import {
  List, Datagrid, TextField, DateField, FunctionField,
  Create, Edit, SimpleForm, NumberInput, SelectInput,
  Show, SimpleShowLayout,
  required, EditButton, ShowButton, DeleteButton,
  useRecordContext, useGetList, Link,
} from 'react-admin';
import Typography from '@mui/material/Typography';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableBody from '@mui/material/TableBody';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';

const roleChoices = [
  { id: 'student', name: 'Студент' },
  { id: 'teacher', name: 'Преподаватель' },
];

const subStatusChoices = [
  { id: 'inactive', name: 'Неактивна' },
  { id: 'active', name: 'Активна' },
  { id: 'expired', name: 'Истекла' },
];

const TeacherStudentsList = () => {
  const record = useRecordContext();
  const { data, isLoading } = useGetList('teacher_students', {
    pagination: { page: 1, perPage: 1000 },
    sort: { field: 'telegram_id', order: 'ASC' },
    filter: { teacher_id: record?.id },
  });

  if (!record || record.role !== 'teacher') return null;

  return (
    <div style={{ marginTop: 24 }}>
      <Typography variant="h6" gutterBottom>
        Студенты, присоединившиеся по ссылке
      </Typography>
      {isLoading ? (
        <Typography variant="body2">Загрузка...</Typography>
      ) : !data || data.length === 0 ? (
        <Typography variant="body2" color="textSecondary">Нет присоединившихся студентов</Typography>
      ) : (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Telegram ID</TableCell>
              <TableCell>Имя пользователя</TableCell>
              <TableCell>Подписка</TableCell>
              <TableCell>Зарегистрирован</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data.map((student: any) => (
              <TableRow key={student.telegram_id}>
                <TableCell>
                  <Link to={`/users/${student.telegram_id}`}>
                    {student.telegram_id}
                  </Link>
                </TableCell>
                <TableCell>{student.username ? `@${student.username}` : '—'}</TableCell>
                <TableCell>{student.subscription_status}</TableCell>
                <TableCell>{student.created_at ? new Date(student.created_at).toLocaleDateString() : '—'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
};

export const UserList = () => (
  <List sort={{ field: 'telegram_id', order: 'ASC' }}>
    <Datagrid bulkActionButtons={false}>
      <TextField source="telegram_id" label="Telegram ID" />
      <FunctionField
        label="Имя пользователя"
        render={(record: any) => record.username ? `@${record.username}` : '—'}
      />
      <TextField source="role" label="Роль" />
      <TextField source="subscription_status" label="Подписка" />
      <DateField source="created_at" label="Зарегистрирован" />
      <ShowButton />
      <EditButton />
      <DeleteButton mutationMode="pessimistic" />
    </Datagrid>
  </List>
);

export const UserShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="telegram_id" label="Telegram ID" />
      <FunctionField
        label="Имя пользователя"
        render={(record: any) => record.username ? `@${record.username}` : '—'}
      />
      <TextField source="role" label="Роль" />
      <TextField source="subscription_status" label="Подписка" />
      <DateField source="created_at" label="Зарегистрирован" />
      <TeacherStudentsList />
    </SimpleShowLayout>
  </Show>
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
