package com.rancy21.flowtask.ui.inboxeditor

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rancy21.flowtask.data.entity.InboxEntity
import com.rancy21.flowtask.data.repository.InboxRepository
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import java.util.UUID

class InboxEditorViewModel(
    private val inboxRepository: InboxRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(InboxEditorUiState())
    val uiState: StateFlow<InboxEditorUiState> = _uiState.asStateFlow()

    fun reset() {
        _uiState.value = InboxEditorUiState()
    }

    fun loadItem(itemId: String) {
        viewModelScope.launch {
            inboxRepository.getById(itemId)?.let { item ->
                _uiState.update {
                    it.copy(
                        title = item.title,
                        description = item.description ?: "",
                        isNew = false,
                        itemId = item.id,
                    )
                }
            }
        }
    }

    fun setTitle(title: String) {
        _uiState.update { it.copy(title = title) }
    }

    fun setDescription(description: String) {
        _uiState.update { it.copy(description = description) }
    }

    fun save(): Boolean {
        val state = _uiState.value
        if (state.title.isBlank()) return false

        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true) }

            val item = if (state.isNew) {
                InboxEntity(
                    id = UUID.randomUUID().toString(),
                    title = state.title.trim(),
                    description = state.description.trim().ifBlank { null },
                    createdAt = java.time.Instant.now().toString(),
                )
            } else {
                inboxRepository.getById(state.itemId!!)?.copy(
                    title = state.title.trim(),
                    description = state.description.trim().ifBlank { null },
                ) ?: return@launch
            }

            inboxRepository.save(item)
        }
        return true
    }
}
