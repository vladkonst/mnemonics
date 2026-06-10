import { useRecordContext, Link } from 'react-admin';

interface Props {
  source: string;
  reference: string;
  label?: string;
}

const LinkedIdField = ({ source, reference }: Props) => {
  const record = useRecordContext();
  if (!record) return null;
  const id = record[source];
  if (id == null) return <span style={{ color: 'rgba(0,0,0,0.3)' }}>—</span>;
  return (
    <Link to={`/${reference}/${id}`} onClick={(e: React.MouseEvent) => e.stopPropagation()}>
      {String(id)}
    </Link>
  );
};

export default LinkedIdField;
