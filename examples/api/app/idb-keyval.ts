// reference: https://github.com/jakearchibald/idb-keyval/blob/main/src/index.ts

function promisifyRequest<T = undefined>(
    request: IDBRequest<T> | IDBTransaction,
): Promise<T> {
    return new Promise<T>((resolve, reject) => {
        // @ts-ignore - file size hacks
        request.oncomplete = request.onsuccess = () => resolve(request.result);
        // @ts-ignore - file size hacks
        request.onabort = request.onerror = () => reject(request.error);
    });
}

function createStore(dbName: string, storeName: string): UseStore {
    const request = indexedDB.open(dbName);
    request.onupgradeneeded = () => request.result.createObjectStore(storeName);
    const dbp = promisifyRequest(request);

    return (txMode, callback) =>
        dbp.then((db) =>
            callback(db.transaction(storeName, txMode).objectStore(storeName)),
        );
}

type UseStore = <T>(
    txMode: IDBTransactionMode,
    callback: (store: IDBObjectStore) => T | PromiseLike<T>,
) => Promise<T>;

let defaultGetStoreFunc: UseStore | undefined;

function defaultGetStore() {
    if (!defaultGetStoreFunc) {
        defaultGetStoreFunc = createStore('keyval-store', 'keyval');
    }
    return defaultGetStoreFunc;
}

/**
 * Get a value by its key.
 *
 * @param key
 * @param customStore Method to get a custom store. Use with caution (see the docs).
 */
function get<T = any>(
    key: IDBValidKey,
    customStore = defaultGetStore(),
): Promise<T | undefined> {
    return customStore('readonly', (store) => promisifyRequest(store.get(key)));
}

/**
 * Set a value with a key.
 *
 * @param key
 * @param value
 * @param customStore Method to get a custom store. Use with caution (see the docs).
 */
function set(
    key: IDBValidKey,
    value: any,
    customStore = defaultGetStore(),
): Promise<void> {
    return customStore('readwrite', (store) => {
        store.put(value, key);
        return promisifyRequest(store.transaction);
    });
}

/**
 * Delete a particular key from the store.
 *
 * @param key
 * @param customStore Method to get a custom store. Use with caution (see the docs).
 */
function del(
    key: IDBValidKey,
    customStore = defaultGetStore(),
): Promise<void> {
    return customStore('readwrite', (store) => {
        store.delete(key);
        return promisifyRequest(store.transaction);
    });
}