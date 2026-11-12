import { useEffect, useState } from 'react';
import { subscribeToFavorites, subscribeToWatchlist } from '../services/firestore';

export function useCollections(userId) {
  const [favorites, setFavorites] = useState(null);
  const [watchlist, setWatchlist] = useState(null);
  const [error, setError] = useState('');

  useEffect(() => {
    const handleError = () => setError('Could not load your collections.');
    const stopFavorites = subscribeToFavorites(userId, setFavorites, handleError);
    const stopWatchlist = subscribeToWatchlist(userId, setWatchlist, handleError);
    return () => {
      stopFavorites();
      stopWatchlist();
    };
  }, [userId]);

  return { favorites, watchlist, error };
}
