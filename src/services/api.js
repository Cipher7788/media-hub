// src/services/api.js

const TMDB_BASE_URL = 'https://api.themoviedb.org/3';
const JIKAN_BASE_URL = 'https://api.jikan.moe/v4';
const COMICVINE_BASE_URL = 'https://comicvine.gamespot.com/api';

// Search movies via TMDB API
export async function searchMovies(query) {
  const apiKey = process.env.REACT_APP_TMDB_API_KEY;
  const url = `${TMDB_BASE_URL}/search/movie?api_key=${apiKey}&query=${encodeURIComponent(query)}`;
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`TMDB API error: ${response.status}`);
  }
  const data = await response.json();
  return data.results || [];
}

// Search anime via Jikan API (no key required)
export async function searchAnime(query) {
  const url = `${JIKAN_BASE_URL}/anime?q=${encodeURIComponent(query)}&limit=20`;
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Jikan API error: ${response.status}`);
  }
  const data = await response.json();
  return data.data || [];
}

// Search comics via Comic Vine API
export async function searchComics(query) {
  const apiKey = process.env.REACT_APP_COMICVINE_API_KEY;
  // Note: API key is sent as a query parameter per Comic Vine API requirements
  const url = `${COMICVINE_BASE_URL}/search/?api_key=${apiKey}&query=${encodeURIComponent(query)}&resources=issue&format=json&limit=20`;
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Comic Vine API error: ${response.status}`);
  }
  const data = await response.json();
  return data.results || [];
}

// Search books via Open Library API (no key required)
export async function searchBooks(query) {
  const url = `https://openlibrary.org/search.json?q=${encodeURIComponent(query)}&limit=20`;
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Open Library API error: ${response.status}`);
  }
  const data = await response.json();
  return data.docs || [];
}
