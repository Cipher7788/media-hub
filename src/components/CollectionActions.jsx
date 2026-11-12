import React, { useState } from 'react';
import {
  addFavorite,
  addToWatchlist,
  removeFavorite,
  removeFromWatchlist,
} from '../services/firestore';

export function CollectionActions({ userId, item, favorites, watchlist }) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');

  const favorite = favorites.find((entry) => entry.mediaId === item.mediaId);
  const queued = watchlist.find((entry) => entry.mediaId === item.mediaId);

  async function toggle(entry, add, remove) {
    setPending(true);
    setError('');
    try {
      if (entry) await remove(entry.id);
      else await add(userId, item);
    } catch {
      setError('Could not update your collection. Try again.');
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="collection-actions">
      <button type="button" disabled={pending} aria-pressed={Boolean(favorite)}
        onClick={() => toggle(favorite, addFavorite, removeFavorite)}>
        {favorite ? 'Remove favorite' : 'Add to favorites'}
      </button>
      <button type="button" disabled={pending} aria-pressed={Boolean(queued)}
        onClick={() => toggle(queued, addToWatchlist, removeFromWatchlist)}>
        {queued ? 'Remove from watchlist' : 'Add to watchlist'}
      </button>
      {error && <p role="alert">{error}</p>}
    </div>
  );
}
