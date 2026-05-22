package com.rancy21.flowtask.data.repository

import com.rancy21.flowtask.data.dao.InboxDao
import com.rancy21.flowtask.data.entity.InboxEntity
import kotlinx.coroutines.flow.Flow

class InboxRepository(private val inboxDao: InboxDao) {
    fun getAll(): Flow<List<InboxEntity>> = inboxDao.getAll()

    suspend fun getById(id: String): InboxEntity? = inboxDao.getById(id)

    suspend fun save(item: InboxEntity) = inboxDao.insert(item)

    suspend fun update(item: InboxEntity) = inboxDao.update(item)

    suspend fun delete(item: InboxEntity) = inboxDao.delete(item)

    suspend fun deleteById(id: String) = inboxDao.deleteById(id)

    suspend fun upsert(item: InboxEntity) {
        if (inboxDao.getById(item.id) != null) {
            inboxDao.update(item)
        } else {
            inboxDao.insert(item)
        }
    }
}
