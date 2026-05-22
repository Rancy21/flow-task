package com.rancy21.flowtask.ui.inbox

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rancy21.flowtask.data.entity.InboxEntity
import com.rancy21.flowtask.data.repository.InboxRepository
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import java.util.UUID

class InboxViewModel(
    private val inboxRepository: InboxRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(InboxUiState())
    val uiState: StateFlow<InboxUiState> = _uiState.asStateFlow()

    init {
        loadAll()
    }

    fun loadAll() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            inboxRepository.getAll().collect { items ->
                _uiState.update { it.copy(items = items, isLoading = false) }
            }
        }
    }

    fun addItem(title: String, description: String) {
        viewModelScope.launch {
            val item = InboxEntity(
                id = UUID.randomUUID().toString(),
                title = title,
                description = description.ifBlank { null },
                createdAt = java.time.Instant.now().toString(),
            )
            inboxRepository.save(item)
        }
    }

    fun delete(item: InboxEntity) {
        viewModelScope.launch {
            inboxRepository.delete(item)
        }
    }
}
