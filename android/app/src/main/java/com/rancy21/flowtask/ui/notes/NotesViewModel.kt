package com.rancy21.flowtask.ui.notes

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rancy21.flowtask.data.entity.NoteEntity
import com.rancy21.flowtask.data.repository.NoteRepository
import com.rancy21.flowtask.data.repository.TaskRepository
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

class NotesViewModel(
    private val noteRepository: NoteRepository,
    private val taskRepository: TaskRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(NotesUiState())
    val uiState: StateFlow<NotesUiState> = _uiState.asStateFlow()

    init {
        loadNotes()
    }

    fun loadNotes() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            noteRepository.getAllNotes().collect { notes ->
                val withTasks = notes.map { note ->
                    val task = taskRepository.getById(note.taskId)
                    NoteWithTask(
                        note = note,
                        taskTitle = task?.title ?: "Deleted task",
                    )
                }
                _uiState.update { it.copy(notes = withTasks, isLoading = false) }
            }
        }
    }

    fun delete(note: NoteEntity) {
        viewModelScope.launch {
            noteRepository.delete(note)
        }
    }
}
