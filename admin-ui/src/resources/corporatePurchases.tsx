import {
  List,
  Datagrid,
  TextField,
  NumberField,
  DateField,
  ShowButton,
  Show,
  SimpleShowLayout,
  ArrayField,
  Button,
  useRecordContext,
} from 'react-admin';
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf';

const PdfDownloadButton = () => {
  const record = useRecordContext();
  if (!record) return null;
  const handleClick = () => {
    const url = `/api/v1/admin/corporate-purchases/${record.id}/pdf`;
    const token = localStorage.getItem('admin_token') || '';
    fetch(url, { headers: { 'X-Admin-Token': token } })
      .then((res) => res.blob())
      .then((blob) => {
        const a = document.createElement('a');
        a.href = URL.createObjectURL(blob);
        a.download = `purchase_${record.id}.pdf`;
        a.click();
      });
  };
  return (
    <Button label="PDF" onClick={handleClick} startIcon={<PictureAsPdfIcon />} />
  );
};

export const CorporatePurchaseList = () => (
  <List sort={{ field: 'created_at', order: 'DESC' }} exporter={false}>
    <Datagrid bulkActionButtons={false}>
      <TextField source="payment_id" label="ID платежа" />
      <TextField source="manager_id" label="Менеджер (TG ID)" />
      <NumberField source="groups_count" label="Групп" />
      <NumberField source="semesters" label="Семестров" />
      <NumberField source="total_amount" label="Сумма (₽)" />
      <DateField source="created_at" label="Создана" showTime />
      <ShowButton />
      <PdfDownloadButton />
    </Datagrid>
  </List>
);

export const CorporatePurchaseShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="payment_id" label="ID платежа" />
      <TextField source="manager_id" label="Менеджер (Telegram ID)" />
      <NumberField source="groups_count" label="Количество групп" />
      <NumberField source="semesters" label="Семестров" />
      <NumberField source="total_amount" label="Сумма (₽)" />
      <DateField source="created_at" label="Создана" showTime />
      <ArrayField source="groups">
        <Datagrid bulkActionButtons={false}>
          <TextField source="id" label="ID группы" />
          <TextField source="name" label="Название" />
          <TextField source="teacher_id" label="Преподаватель (TG ID)" />
          <TextField source="teacher_join_code" label="Код преподавателя" />
          <TextField source="student_link_id" label="ID студенческой ссылки" />
        </Datagrid>
      </ArrayField>
    </SimpleShowLayout>
  </Show>
);
