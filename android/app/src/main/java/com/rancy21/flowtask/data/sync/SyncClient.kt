package com.rancy21.flowtask.data.sync

import android.util.Log
import com.rancy21.flowtask.data.entity.InboxEntity
import com.rancy21.flowtask.data.entity.NoteEntity
import com.rancy21.flowtask.data.entity.TaskEntity
import com.rancy21.flowtask.data.repository.InboxRepository
import com.rancy21.flowtask.data.repository.NoteRepository
import com.rancy21.flowtask.data.repository.TaskRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.util.concurrent.TimeUnit

class SyncClient(
    private val taskRepo: TaskRepository,
    private val noteRepo: NoteRepository,
    private val inboxRepo: InboxRepository,
) {
    private val client = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(15, TimeUnit.SECONDS)
        .build()

    private val json = Json { ignoreUnknownKeys = true; isLenient = true }

    // ── Pull ──────────────────────────────────────────────────────────────────

    suspend fun pullAll() {
        withContext(Dispatchers.IO) {
            pullTasks()
            pullNotes()
            pullInbox()
        }
    }

    private suspend fun pullTasks() {
        val url = "$BASE_URL/rest/v1/tasks?select=*&order=updated_at.asc"
        val tasks = get<List<SupabaseTask>>(url) ?: return
        for (st in tasks) {
            taskRepo.upsert(st.toEntity())
        }
    }

    private suspend fun pullNotes() {
        val url = "$BASE_URL/rest/v1/notes?select=*&order=updated_at.asc"
        val notes = get<List<SupabaseNote>>(url) ?: return
        for (sn in notes) {
            noteRepo.upsert(sn.toEntity())
        }
    }

    private suspend fun pullInbox() {
        val url = "$BASE_URL/rest/v1/inbox?select=*&order=updated_at.asc"
        val items = get<List<SupabaseInboxItem>>(url) ?: return
        for (si in items) {
            inboxRepo.upsert(si.toEntity())
        }
    }

    // ── Push ──────────────────────────────────────────────────────────────────

    suspend fun pushTask(task: TaskEntity) {
        withContext(Dispatchers.IO) {
            val body = json.encodeToString(SupabaseTask.serializer(), SupabaseTask.fromEntity(task))
            // Try PATCH first, POST if not found
            val patchUrl = "$BASE_URL/rest/v1/tasks?id=eq.${task.id}"
            val patchResp = request("PATCH", patchUrl, body)
            if (patchResp in setOf(404, 406)) {
                request("POST", "$BASE_URL/rest/v1/tasks", body)
            }
        }
    }

    suspend fun pushNote(note: NoteEntity) {
        withContext(Dispatchers.IO) {
            val body = json.encodeToString(SupabaseNote.serializer(), SupabaseNote.fromEntity(note))
            val patchResp = request("PATCH", "$BASE_URL/rest/v1/notes?id=eq.${note.id}", body)
            if (patchResp in setOf(404, 406)) {
                request("POST", "$BASE_URL/rest/v1/notes", body)
            }
        }
    }

    suspend fun pushInboxItem(item: InboxEntity) {
        withContext(Dispatchers.IO) {
            val body = json.encodeToString(SupabaseInboxItem.serializer(), SupabaseInboxItem.fromEntity(item))
            val patchResp = request("PATCH", "$BASE_URL/rest/v1/inbox?id=eq.${item.id}", body)
            if (patchResp in setOf(404, 406)) {
                request("POST", "$BASE_URL/rest/v1/inbox", body)
            }
        }
    }

    suspend fun deleteTask(id: String) {
        withContext(Dispatchers.IO) {
            request("DELETE", "$BASE_URL/rest/v1/tasks?id=eq.$id", null)
        }
    }

    suspend fun deleteNote(id: String) {
        withContext(Dispatchers.IO) {
            request("DELETE", "$BASE_URL/rest/v1/notes?id=eq.$id", null)
        }
    }

    suspend fun deleteInboxItem(id: String) {
        withContext(Dispatchers.IO) {
            request("DELETE", "$BASE_URL/rest/v1/inbox?id=eq.$id", null)
        }
    }

    // ── HTTP helpers ──────────────────────────────────────────────────────────

    private inline fun <reified T> get(url: String): T? {
        return try {
            val req = Request.Builder().url(url).get()
                .header("apikey", API_KEY)
                .header("Authorization", "Bearer $API_KEY")
                .build()
            val resp = client.newCall(req).execute()
            val body = resp.body?.string() ?: return null
            if (!resp.isSuccessful) {
                Log.w("SyncClient", "GET $url → ${resp.code}: $body")
                return null
            }
            json.decodeFromString(body)
        } catch (e: Exception) {
            Log.e("SyncClient", "GET $url failed", e)
            null
        }
    }

    private fun request(method: String, url: String, body: String?): Int {
        return try {
            val reqBuilder = Request.Builder().url(url)
            when (method) {
                "PATCH" -> reqBuilder.patch(body!!.toRequestBody(JSON))
                "POST" -> reqBuilder.post(body!!.toRequestBody(JSON))
                "DELETE" -> reqBuilder.delete()
            }
            reqBuilder
                .header("apikey", API_KEY)
                .header("Authorization", "Bearer $API_KEY")
                .header("Content-Type", "application/json")
                .header("Prefer", "return=minimal")
            val resp = client.newCall(reqBuilder.build()).execute()
            if (resp.code >= 400) {
                Log.w("SyncClient", "$method $url → ${resp.code}: ${resp.body?.string()}")
            }
            resp.code
        } catch (e: Exception) {
            Log.e("SyncClient", "$method $url failed", e)
            -1
        }
    }

    companion object {
        private const val BASE_URL = "https://ykksgiyweklxbrfoomwa.supabase.co"
        private const val API_KEY = "sb_publishable_8JsM8svXjt1-yagX9M0n_w_jN7OWZBZ"
        private val JSON = "application/json".toMediaType()
    }
}

// ── Supabase JSON types ───────────────────────────────────────────────────────

@Serializable
data class SupabaseTask(
    val id: String,
    val title: String,
    val description: String? = null,
    val priority: String,
    val status: String,
    val scheduled_date: String? = null,
    val created_at: String,
    val completed_at: String? = null,
    val updated_at: String = "",
) {
    fun toEntity() = TaskEntity(id, title, description, priority, status, scheduled_date, created_at, completed_at, updated_at)

    companion object {
        fun fromEntity(t: TaskEntity) = SupabaseTask(t.id, t.title, t.description, t.priority, t.status, t.scheduledDate, t.createdAt, t.completedAt, t.updatedAt)
    }
}

@Serializable
data class SupabaseNote(
    val id: String,
    val task_id: String,
    val content: String,
    val created_at: String,
    val updated_at: String = "",
) {
    fun toEntity() = NoteEntity(id, task_id, content, created_at, updated_at)

    companion object {
        fun fromEntity(n: NoteEntity) = SupabaseNote(n.id, n.taskId, n.content, n.createdAt, n.updatedAt)
    }
}

@Serializable
data class SupabaseInboxItem(
    val id: String,
    val title: String,
    val description: String? = null,
    val created_at: String,
    val updated_at: String = "",
) {
    fun toEntity() = InboxEntity(id, title, description, created_at, updated_at)

    companion object {
        fun fromEntity(i: InboxEntity) = SupabaseInboxItem(i.id, i.title, i.description, i.createdAt, i.updatedAt)
    }
}
