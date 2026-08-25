import axios from "axios"
const api = axios.create({ baseURL: import.meta.env.VITE_API_BASE || "http://localhost:8080/api/v1" })
api.interceptors.request.use(config => { const token = localStorage.getItem("kac_access_token"); if (token) config.headers.Authorization = `Bearer ${token}`; return config })
export async function login(email: string, password: string) { const { data } = await api.post("/auth/login", { email, password }); localStorage.setItem("kac_access_token", data.access_token); return data }
export const workspaces = () => api.get("/workspaces").then(r => r.data.items || [])
export const documents = (workspaceId: string, q = "") => api.get("/documents", { params: { workspace_id: workspaceId, q } }).then(r => r.data.items || [])
export const createWorkspace = (payload: any) => api.post("/workspaces", payload).then(r => r.data)
export const createDocument = (payload: any) => api.post("/documents", payload).then(r => r.data)
export const saveDraft = (id: string, body: string, expectedVersion: number) => api.patch(`/documents/${id}/draft`, { body, expected_version: expectedVersion }).then(r => r.data)
export const documentDetail = (id: string) => api.get(`/documents/${id}`).then(r => r.data)
export const versions = (id: string) => api.get(`/documents/${id}/versions`).then(r => r.data.items || [])
export const comments = (id: string) => api.get(`/documents/${id}/comments`).then(r => r.data.items || [])
export const addComment = (id: string, body: string) => api.post(`/documents/${id}/comments`, { body }).then(r => r.data)
export const search = (q: string) => api.get("/search", { params: { q } }).then(r => r.data.items || [])
export const recycle = () => api.get("/recycle-bin").then(r => r.data.items || [])
export const report = (workspaceId: string) => api.get("/reports", { params: { workspace_id: workspaceId } }).then(r => r.data)
export default api
