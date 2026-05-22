package com.rancy21.flowtask.ui.editor

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rancy21.flowtask.data.entity.NoteEntity
import com.rancy21.flowtask.data.entity.TaskEntity
import com.rancy21.flowtask.data.repository.NoteRepository
import com.rancy21.flowtask.data.repository.TaskRepository
import com.rancy21.flowtask.data.sync.SyncClient
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import java.util.UUID

class TaskEditorViewModel(
    private val taskRepository: TaskRepository,
    private val noteRepository: NoteRepository,
    private val syncClient: SyncClient,
) : ViewModel() {

    private val _uiState = MutableStateFlow(TaskEditorUiState())
    val uiState: StateFlow<TaskEditorUiState> = _uiState.asStateFlow()

    fun reset() {
        _uiState.value = TaskEditorUiState()
    }

    fun loadTask(taskId: String) {
        viewModelScope.launch {
            taskRepository.getById(taskId)?.let { task ->
                _uiState.update {
                    it.copy(
                        title = task.title,
                        description = task.description ?: "",
                        priority = task.priority,
                        scheduledDate = task.scheduledDate,
                        isNew = false,
                        taskId = task.id,
                    )
                }
            }
            loadNotes(taskId)
        }
    }

    private suspend fun loadNotes(taskId: String) {
        noteRepository.getNotesForTask(taskId).collect { notes ->
            _uiState.update { it.copy(notes = notes) }
        }
    }

    fun setTitle(title: String) {
        _uiState.update { it.copy(title = title) }
    }

    fun setDescription(description: String) {
        _uiState.update { it.copy(description = description) }
    }

    fun setPriority(priority: String) {
        _uiState.update { it.copy(priority = priority) }
    }

    fun setScheduledDate(date: String?) {
        _uiState.update { it.copy(scheduledDate = date) }
    }

    fun setNewNoteContent(content: String) {
        _uiState.update { it.copy(newNoteContent = content) }
    }

    fun addNote() {
        val state = _uiState.value
        val content = state.newNoteContent.trim()
        if (content.isBlank()) return
        val taskId = state.taskId
        if (taskId == null) {
            _uiState.update { it.copy(newNoteContent = "") }
            return
        }

        viewModelScope.launch {
            val note = NoteEntity(
                id = UUID.randomUUID().toString(),
                taskId = taskId,
                content = content,
                createdAt = java.time.Instant.now().toString(),
            )
            noteRepository.save(note)
            syncClient.pushNote(note)
            _uiState.update { it.copy(newNoteContent = "") }
        }
    }

    fun deleteNote(note: NoteEntity) {
        viewModelScope.launch {
            noteRepository.delete(note)
        }
    }

    fun save(): Boolean {
        val state = _uiState.value
        if (state.title.isBlank()) return false

        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true) }

            val now = java.time.Instant.now().toString()
            val status = if (state.scheduledDate != null) "SCHEDULED" else "INBOX"

            val task = if (state.isNew) {
                TaskEntity(
                    id = UUID.randomUUID().toString(),
                    title = state.title.trim(),
                    description = state.description.trim().ifBlank { null },
                    priority = state.priority,
                    status = status,
                    scheduledDate = state.scheduledDate,
                    createdAt = now,
                    completedAt = null,
                )
            } else {
                taskRepository.getById(state.taskId!!)?.copy(
                    title = state.title.trim(),
                    description = state.description.trim().ifBlank { null },
                    priority = state.priority,
                    status = status,
                    scheduledDate = state.scheduledDate,
                ) ?: return@launch
            }

            taskRepository.save(task)

            // Push to Supabase
            syncClient.pushTask(task)

            val pendingNote = state.newNoteContent.trim()
            if (pendingNote.isNotBlank()) {
                noteRepository.save(
                    NoteEntity(
                        id = UUID.randomUUID().toString(),
                        taskId = task.id,
                        content = pendingNote,
                        createdAt = now,
                    )
                )
            }

            _uiState.update { it.copy(isSaving = false, newNoteContent = "") }
        }
        return true
    }
}
