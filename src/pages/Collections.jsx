import React, { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useCollections } from '../hooks/useCollections';
import { removeFavorite, removeFromWatchlist } from '../services/firestore';
import './Collections.css';

function CollectionEntry({ item, remove }) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');

  async function handleRemove() {
    setPending(true);
    setError('');
    try {
      await remove(item.id);
    } catch {
      setError('Could not remove this item. Try again.');
    } finally {
      setPending(false);
    }
  }

  return (
    <li className="collection-entry">
      <span>{item.title || 'Untitled item'}</span>
      <button type="button" disabled={pending} onClick={handleRemove}
        aria-label={`Remove ${item.title || 'item'}`}>
        Remove
      </button>
      {error && <p role="alert">{error}</p>}
    </li>
  );
}

function CollectionList({ title, items, remove }) {
  return (
    <section>
      <h2>{title}</h2>
      {items.length ? (
        <ul>{items.map((item) => (
          <CollectionEntry key={item.id} item={item} remove={remove} />
        ))}</ul>
      ) : <p>No saved items.</p>}
    </section>
  );
}

function Dashboard({ userId }) {
  const { favorites, watchlist, error } = useCollections(userId);

  return (
    <main className="collections">
      <h1>My collections</h1>
      {error && <p role="alert">{error}</p>}
      {!error && (!favorites || !watchlist) && <p role="status">Loading collections…</p>}
      {favorites && <CollectionList title="Favorites" items={favorites} remove={removeFavorite} />}
      {watchlist && <CollectionList title="Watchlist" items={watchlist} remove={removeFromWatchlist} />}
    </main>
  );
}

export function Collections() {
  const { currentUser } = useAuth();
  if (!currentUser) return <Navigate to="/login" state={{ returnTo: '/collections' }} replace />;
  return <Dashboard key={currentUser.uid} userId={currentUser.uid} />;
}
