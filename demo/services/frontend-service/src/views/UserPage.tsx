import { Layout } from "./Layout.js";

export interface User {
  id: string;
  name: string;
  email: string;
  phone: string;
  address: string;
  avatarUrl: string;
  bio: string;
  joinedAt: string;
}

export function UserPage({ user }: { user: User }) {
  return (
    <Layout title={`SHOP - ${user.name}`}>
      <h1 className="page-title">Профиль</h1>
      <div className="profile">
        <img className="profile-avatar" src={user.avatarUrl} alt={user.name} />
        <div className="profile-body">
          <div className="profile-name">{user.name}</div>
          {user.bio ? <p className="profile-bio">{user.bio}</p> : null}

          <dl className="profile-fields">
            <Field label="ID">{user.id}</Field>
            <Field label="Email">
              <a href={`mailto:${user.email}`}>{user.email}</a>
            </Field>
            {user.phone ? (
              <Field label="Телефон">
                <a href={`tel:${user.phone.replace(/\s+/g, "")}`}>{user.phone}</a>
              </Field>
            ) : null}
            {user.address ? <Field label="Адрес">{user.address}</Field> : null}
            {user.joinedAt ? <Field label="С нами с">{user.joinedAt}</Field> : null}
          </dl>
        </div>
      </div>
    </Layout>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <>
      <dt>{label}</dt>
      <dd>{children}</dd>
    </>
  );
}
