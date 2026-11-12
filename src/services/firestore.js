// src/services/firestore.js
import {
  collection,
  doc,
  getDocs,
  addDoc,
  deleteDoc,
  updateDoc,
  query,
  where,
} from 'firebase/firestore';
import { db } from '../config/firebase';

// --- Favorites ---

export async function getFavorites(userId) {
  const q = query(collection(db, 'favorites'), where('userId', '==', userId));
  const snapshot = await getDocs(q);
  return snapshot.docs.map((d) => ({ ...d.data(), id: d.id }));
}

export async function addFavorite(userId, item) {
  return addDoc(collection(db, 'favorites'), { ...item, userId });
}

export async function removeFavorite(favoriteId) {
  return deleteDoc(doc(db, 'favorites', favoriteId));
}

// --- Watchlist ---

export async function getWatchlist(userId) {
  const q = query(collection(db, 'watchlist'), where('userId', '==', userId));
  const snapshot = await getDocs(q);
  return snapshot.docs.map((d) => ({ ...d.data(), id: d.id }));
}

export async function addToWatchlist(userId, item) {
  return addDoc(collection(db, 'watchlist'), { ...item, userId });
}

export async function removeFromWatchlist(watchlistId) {
  return deleteDoc(doc(db, 'watchlist', watchlistId));
}

// --- Reviews ---

export async function getReviews(mediaId) {
  const q = query(collection(db, 'reviews'), where('mediaId', '==', mediaId));
  const snapshot = await getDocs(q);
  return snapshot.docs.map((d) => ({ ...d.data(), id: d.id }));
}

export async function addReview(userId, mediaId, reviewData) {
  return addDoc(collection(db, 'reviews'), { ...reviewData, userId, mediaId });
}

export async function updateReview(reviewId, reviewData) {
  return updateDoc(doc(db, 'reviews', reviewId), reviewData);
}

export async function deleteReview(reviewId) {
  return deleteDoc(doc(db, 'reviews', reviewId));
}

