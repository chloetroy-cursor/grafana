import { useState } from 'react';

import { css } from '@emotion/css';

import { dateTime, type GrafanaTheme2 } from '@grafana/data';
import { t } from '@grafana/i18n';
import { Alert, Button, Input, TextArea, useStyles2 } from '@grafana/ui';
import {
  useCreateDashboardHandoffNoteMutation,
  useDeleteDashboardHandoffNoteMutation,
  useGetDashboardHandoffNotesQuery,
} from 'app/features/dashboard/api/handoffNotesApi';

import { type DashboardScene } from './DashboardScene';

interface Props {
  dashboard: DashboardScene;
}

export function DashboardHandoffNote({ dashboard }: Props) {
  const styles = useStyles2(getStyles);
  const { uid, meta } = dashboard.state;
  const canWrite = Boolean(meta.canEdit);
  const skip = !uid || Boolean(meta.isSnapshot || meta.isEmbedded);
  const { data: notes = [], isLoading } = useGetDashboardHandoffNotesQuery(uid!, { skip });
  const [createNote, createState] = useCreateDashboardHandoffNoteMutation();
  const [deleteNote, deleteState] = useDeleteDashboardHandoffNoteMutation();
  const [isEditing, setIsEditing] = useState(false);
  const [editingNoteId, setEditingNoteId] = useState<number>();
  const [text, setText] = useState('');
  const [mentions, setMentions] = useState('');
  const [error, setError] = useState<string>();
  const latest = notes[0];

  if (skip || isLoading || (!latest && !canWrite)) {
    return null;
  }

  const startCreate = () => {
    setEditingNoteId(undefined);
    setText('');
    setMentions('');
    setError(undefined);
    setIsEditing(true);
  };

  const startEdit = () => {
    if (!latest) {
      return;
    }
    setEditingNoteId(latest.id);
    setText(latest.text);
    setMentions(latest.mentions.map((mention) => mention.value).join(', '));
    setError(undefined);
    setIsEditing(true);
  };

  const cancelEdit = () => {
    setIsEditing(false);
    setEditingNoteId(undefined);
    setError(undefined);
  };

  const save = async () => {
    if (!text.trim()) {
      setError(t('dashboard.handoff-notes.text-required', 'Enter a handoff note before saving.'));
      return;
    }

    try {
      setError(undefined);
      await createNote({
        dashboardUid: uid!,
        text,
        mentions: mentions
          .split(',')
          .map((mention) => mention.trim())
          .filter(Boolean),
      }).unwrap();
      if (editingNoteId) {
        await deleteNote({ dashboardUid: uid!, id: editingNoteId }).unwrap();
      }
      cancelEdit();
    } catch {
      setError(t('dashboard.handoff-notes.save-error', 'The handoff note could not be saved.'));
    }
  };

  const remove = async () => {
    if (!latest) {
      return;
    }
    try {
      setError(undefined);
      await deleteNote({ dashboardUid: uid!, id: latest.id }).unwrap();
    } catch {
      setError(t('dashboard.handoff-notes.delete-error', 'The handoff note could not be deleted.'));
    }
  };

  return (
    <section className={styles.container} aria-label={t('dashboard.handoff-notes.label', 'Dashboard handoff note')}>
      <div className={styles.heading}>
        <strong>{t('dashboard.handoff-notes.title', 'Handoff note')}</strong>
        {!isEditing && canWrite && (
          <div className={styles.actions}>
            {latest ? (
              <>
                <Button variant="secondary" size="sm" onClick={startEdit}>
                  {t('dashboard.handoff-notes.edit', 'Edit')}
                </Button>
                <Button variant="destructive" size="sm" onClick={remove} disabled={deleteState.isLoading}>
                  {t('dashboard.handoff-notes.delete', 'Delete')}
                </Button>
              </>
            ) : (
              <Button variant="secondary" size="sm" onClick={startCreate}>
                {t('dashboard.handoff-notes.add', 'Add note')}
              </Button>
            )}
          </div>
        )}
      </div>

      {error && <Alert severity="error" title={error} />}

      {isEditing ? (
        <div className={styles.editor}>
          <TextArea
            value={text}
            rows={3}
            maxLength={10000}
            autoFocus
            placeholder={t(
              'dashboard.handoff-notes.placeholder',
              'Leave context for the next on-call engineer. Markdown is supported.'
            )}
            onChange={(event) => setText(event.currentTarget.value)}
          />
          <Input
            value={mentions}
            placeholder={t(
              'dashboard.handoff-notes.mentions-placeholder',
              'Mention logins or emails, separated by commas'
            )}
            onChange={(event) => setMentions(event.currentTarget.value)}
          />
          <div className={styles.actions}>
            <Button size="sm" onClick={save} disabled={createState.isLoading || deleteState.isLoading}>
              {t('dashboard.handoff-notes.save', 'Save note')}
            </Button>
            <Button variant="secondary" size="sm" onClick={cancelEdit}>
              {t('dashboard.handoff-notes.cancel', 'Cancel')}
            </Button>
          </div>
        </div>
      ) : (
        latest && (
          <>
            <div className={styles.noteBody} dangerouslySetInnerHTML={{ __html: latest.html }} />
            <div className={styles.metadata}>
              {latest.authorLogin} · {dateTime(latest.createdAt).fromNow()}
            </div>
          </>
        )
      )}
    </section>
  );
}

const getStyles = (theme: GrafanaTheme2) => ({
  container: css({
    backgroundColor: theme.colors.background.secondary,
    border: `1px solid ${theme.colors.border.weak}`,
    borderColor: theme.colors.border.weak,
    borderRadius: theme.shape.radius.default,
    margin: theme.spacing(0, 2, 1),
    padding: theme.spacing(1, 2),
  }),
  heading: css({
    alignItems: 'center',
    display: 'flex',
    justifyContent: 'space-between',
    marginBottom: theme.spacing(1),
  }),
  actions: css({
    display: 'flex',
    gap: theme.spacing(1),
  }),
  editor: css({
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(1),
  }),
  noteBody: css({
    color: theme.colors.text.primary,
    '& > :last-child': {
      marginBottom: 0,
    },
  }),
  metadata: css({
    color: theme.colors.text.secondary,
    fontSize: theme.typography.bodySmall.fontSize,
    marginTop: theme.spacing(0.5),
  }),
});
