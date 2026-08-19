import { legacyAPI } from 'app/api/clients/legacy';

export interface DashboardHandoffNoteMention {
  value: string;
  userId?: number;
}

export interface DashboardHandoffNote {
  id: number;
  authorId: number;
  authorLogin: string;
  text: string;
  html: string;
  createdAt: string;
  mentions: DashboardHandoffNoteMention[];
}

export const handoffNotesApi = legacyAPI.enhanceEndpoints({ addTagTypes: ['DashboardHandoffNotes'] }).injectEndpoints({
  endpoints: (builder) => ({
    getDashboardHandoffNotes: builder.query<DashboardHandoffNote[], string>({
      query: (dashboardUid) => `/dashboards/uid/${dashboardUid}/handoff-notes`,
      providesTags: (_result, _error, dashboardUid) => [{ type: 'DashboardHandoffNotes', id: dashboardUid }],
    }),
    createDashboardHandoffNote: builder.mutation<
      DashboardHandoffNote,
      { dashboardUid: string; text: string; mentions: string[] }
    >({
      query: ({ dashboardUid, ...body }) => ({
        url: `/dashboards/uid/${dashboardUid}/handoff-notes`,
        method: 'POST',
        body,
      }),
      invalidatesTags: (_result, _error, { dashboardUid }) => [{ type: 'DashboardHandoffNotes', id: dashboardUid }],
    }),
    deleteDashboardHandoffNote: builder.mutation<void, { dashboardUid: string; id: number }>({
      query: ({ dashboardUid, id }) => ({
        url: `/dashboards/uid/${dashboardUid}/handoff-notes/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { dashboardUid }) => [{ type: 'DashboardHandoffNotes', id: dashboardUid }],
    }),
  }),
});

export const {
  useGetDashboardHandoffNotesQuery,
  useCreateDashboardHandoffNoteMutation,
  useDeleteDashboardHandoffNoteMutation,
} = handoffNotesApi;
