// src/services/api.js

const TMDB_BASE_URL = 'https://api.themoviedb.org/3';
const JIKAN_BASE_URL = 'https://api.jikan.moe/v4';
const COMICVINE_BASE_URL = 'https://comicvine.gamespot.com/api';

// Search movies via TMDB API
export async function searchMovies(query, { signal } = {}) {
  const apiKey = process.env.REACT_APP_TMDB_API_KEY;
  if (!apiKey) throw new Error('Movie search requires a configured TMDB API key.');
  const url = `${TMDB_BASE_URL}/search/movie?api_key=${apiKey}&query=${encodeURIComponent(query)}`;
  const response = await fetch(url, { signal });
  if (!response.ok) {
    throw new Error(`TMDB API error: ${response.status}`);
  }
  const data = await response.json();
  return data.results || [];
}

export async function discoverMovies(genre, { signal } = {}) {
  const apiKey = process.env.REACT_APP_TMDB_API_KEY;
  if (!apiKey) throw new Error('Movie browsing requires a configured TMDB API key.');
  const params = new URLSearchParams({
    api_key: apiKey,
    with_genres: genre,
  });
  const response = await fetch(`${TMDB_BASE_URL}/discover/movie?${params}`, { signal });
  if (!response.ok) throw new Error(`TMDB API error: ${response.status}`);
  const data = await response.json();
  return data.results || [];
}

// Search anime via Jikan API (no key required)
export async function searchAnime(query, { signal } = {}) {
  const url = `${JIKAN_BASE_URL}/anime?q=${encodeURIComponent(query)}&limit=20`;
  const response = await fetch(url, { signal });
  if (!response.ok) {
    throw new Error(`Jikan API error: ${response.status}`);
  }
  const data = await response.json();
  return data.data || [];
}

// Search comics via Comic Vine API
export async function searchComics(query, { signal } = {}) {
  const apiKey = process.env.REACT_APP_COMICVINE_API_KEY;
  if (!apiKey) throw new Error('Comics search requires a configured Comic Vine API key.');
  // Note: API key is sent as a query parameter per Comic Vine API requirements
  const url = `${COMICVINE_BASE_URL}/search/?api_key=${apiKey}&query=${encodeURIComponent(query)}&resources=issue&format=json&limit=20`;
  const response = await fetch(url, { signal });
  if (!response.ok) {
    throw new Error(`Comic Vine API error: ${response.status}`);
  }
  const data = await response.json();
  if (!Array.isArray(data.results)) throw new Error('Comic Vine did not return search results.');
  return data.results;
}

// Search books via Open Library API (no key required)
export async function searchBooks(query, { signal } = {}) {
  const url = `https://openlibrary.org/search.json?q=${encodeURIComponent(query)}&limit=20`;
  const response = await fetch(url, { signal });
  if (!response.ok) {
    throw new Error(`Open Library API error: ${response.status}`);
  }
  const data = await response.json();
  return data.docs || [];
}

