// src/pages/Search.jsx
import React, { useState, useEffect } from 'react';
import { Link, useLocation, useSearchParams } from 'react-router-dom';
import { discoverMovies, searchMovies, searchAnime, searchComics, searchBooks } from '../services/api';
import { useAuth } from '../hooks/useAuth';
import { useCollections } from '../hooks/useCollections';
import { Reviews } from '../components/Reviews';
import { CollectionActions } from '../components/CollectionActions';
import { movieGenres } from '../movieGenres';
import './Search.css';

const CATEGORIES = [
  { value: 'movies', label: 'Movies' },
  { value: 'anime', label: 'Anime' },
  { value: 'comics', label: 'Comics' },
  { value: 'books', label: 'Books' },
];

const searchByCategory = new Map([
  ['movies', searchMovies],
  ['anime', searchAnime],
  ['comics', searchComics],
  ['books', searchBooks],
]);

function normalizeResult(item, category) {
  if (category === 'movies') {
    return {
      id: item.id,
      title: item.title,
      image: item.poster_path
        ? `https://image.tmdb.org/t/p/w300${item.poster_path}`
        : null,
      subtitle: item.release_date ? item.release_date.slice(0, 4) : '',
    };
  }
  if (category === 'anime') {
    return {
      id: item.mal_id,
      title: item.title,
      image: item.images?.jpg?.image_url || null,
      subtitle: item.type || '',
    };
  }
  if (category === 'comics') {
    return {
      id: item.id,
      title: item.name,
      image: item.image?.medium_url || null,
      subtitle: item.volume?.name || '',
    };
  }
  if (category === 'books') {
    return {
      id: item.key,
      title: item.title,
      image: item.cover_i
        ? `https://covers.openlibrary.org/b/id/${item.cover_i}-M.jpg`
        : null,
      subtitle: item.author_name ? item.author_name[0] : '',
    };
  }
  return { id: item.id, title: String(item), image: null, subtitle: '' };
}

function SavedResults({ userId, ...props }) {
  const collections = useCollections(userId);
  return <ResultGrid {...props} userId={userId} collections={collections} />;
}

function ResultGrid({ results, category, userId, collections }) {
  const [openReviews, setOpenReviews] = useState(null);

  return (
    <>
      {collections?.error && <p role="alert">{collections.error}</p>}
      <div className="results-grid">
        {results.map((item) => {
          const mediaId = `${category}:${item.id}`;
          return (
            <div key={mediaId} className="result-card">
              {item.image ? (
                <img src={item.image} alt={item.title} loading="lazy" />
              ) : (
                <div className="result-placeholder" aria-hidden="true">🎬</div>
              )}
              <h3>{item.title}</h3>
              {item.subtitle && <p>{item.subtitle}</p>}
              {collections?.favorites && collections?.watchlist && (
                <CollectionActions userId={userId} favorites={collections.favorites}
                  watchlist={collections.watchlist} item={{
                    mediaId,
                    category,
                    title: item.title || 'Untitled item',
                    image: item.image,
                  }} />
              )}
              {userId && (
                <div className="collection-actions">
                  <button type="button" aria-expanded={openReviews === mediaId}
                    onClick={() => setOpenReviews(openReviews === mediaId ? null : mediaId)}>
                    {openReviews === mediaId ? 'Hide reviews' : 'Show reviews'}
                  </button>
                  {openReviews === mediaId && <Reviews mediaId={mediaId} userId={userId} />}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </>
  );
}

export function Search() {
  const { currentUser } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const { key: locationKey } = useLocation();
  const submittedQuery = (searchParams.get('q') || '').trim();
  const urlCategory = searchParams.get('category');
  const submittedCategory = searchByCategory.has(urlCategory) ? urlCategory : 'movies';
  const genre = searchParams.get('genre');
  const genreName = !submittedQuery && submittedCategory === 'movies' ? movieGenres.get(genre) : null;
  const [query, setQuery] = useState(submittedQuery);
  const [category, setCategory] = useState(submittedCategory);
  const [outcome, setOutcome] = useState(null);

  useEffect(() => {
    setQuery(submittedQuery);
    setCategory(submittedCategory);

    if (!submittedQuery && !genreName) return;
    const controller = new AbortController();

    async function runSearch() {
      try {
        const search = searchByCategory.get(submittedCategory);
        const options = { signal: controller.signal };
        const raw = genreName
          ? await discoverMovies(genre, options)
          : await search(submittedQuery, options);
        if (!controller.signal.aborted) {
          const results = raw.map((item) => normalizeResult(item, submittedCategory))
            .filter((item) => item.id != null);
          setOutcome({ locationKey, results });
        }
      } catch (error) {
        if (!controller.signal.aborted) {
          setOutcome({ locationKey, error: error.message, results: [] });
        }
      }
    }

    runSearch();
    return () => controller.abort();
  }, [locationKey, submittedQuery, submittedCategory, genre, genreName]);

  const currentOutcome = outcome?.locationKey === locationKey ? outcome : null;
  const loading = Boolean((submittedQuery || genreName) && !currentOutcome);
  const results = currentOutcome?.results || [];
  const error = currentOutcome?.error;

  function handleSubmit(event) {
    event.preventDefault();
    const trimmedQuery = query.trim();
    setOutcome(null);
    setSearchParams(trimmedQuery ? { q: trimmedQuery, category } : { category });
  }

  return (
    <div className="search">
      <div className="search-container">
        <h1>{genreName ? `${genreName} movies` : 'Search'}</h1>
        <form className="search-form" onSubmit={handleSubmit}>
          <select
            className="search-input"
            aria-label="Media category"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            style={{ flex: '0 0 auto', width: '130px' }}
          >
            {CATEGORIES.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
          </select>
          <input
            className="search-input"
            type="text"
            placeholder="Search..."
            aria-label="Search titles"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <button className="search-button" type="submit">
            Search
          </button>
        </form>

        {loading && <p className="loading" role="status">Loading...</p>}
        {error && <p className="loading" role="alert" style={{ color: '#e53e3e' }}>{error}</p>}

        {!loading && !error && results.length > 0 && (
          currentUser ? (
            <SavedResults key={currentUser.uid} userId={currentUser.uid}
              results={results} category={submittedCategory} />
          ) : (
            <>
              <p><Link to="/login" state={{ returnTo: '/search', search: searchParams.toString() }}>Sign in</Link> to save items or write reviews.</p>
              <ResultGrid results={results} category={submittedCategory} />
            </>
          )
        )}

        {!loading && !error && results.length === 0 && (submittedQuery || genreName) && (
          <p className="loading">No results found.</p>
        )}
      </div>
    </div>
  );
}

