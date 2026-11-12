// src/services/firestore.js
import {
  collection,
  doc,
  getDocs,
  onSnapshot,
  setDoc,
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

export function subscribeToFavorites(userId, onItems, onError) {
  const ownedItems = query(collection(db, 'favorites'), where('userId', '==', userId));
  return onSnapshot(ownedItems, (snapshot) => {
    onItems(snapshot.docs.map((entry) => ({ ...entry.data(), id: entry.id })));
  }, onError);
}

export async function addFavorite(userId, item) {
  if (!item.mediaId) return addDoc(collection(db, 'favorites'), { ...item, userId });
  const id = JSON.stringify([userId, item.mediaId]);
  const entry = doc(db, 'favorites', encodeURIComponent(id));
  await setDoc(entry, { ...item, userId });
  return entry;
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

export function subscribeToWatchlist(userId, onItems, onError) {
  const ownedItems = query(collection(db, 'watchlist'), where('userId', '==', userId));
  return onSnapshot(ownedItems, (snapshot) => {
    onItems(snapshot.docs.map((entry) => ({ ...entry.data(), id: entry.id })));
  }, onError);
}

export async function addToWatchlist(userId, item) {
  if (!item.mediaId) return addDoc(collection(db, 'watchlist'), { ...item, userId });
  const id = JSON.stringify([userId, item.mediaId]);
  const entry = doc(db, 'watchlist', encodeURIComponent(id));
  await setDoc(entry, { ...item, userId });
  return entry;
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

export function subscribeToReviews(mediaId, onReviews, onError) {
  const mediaReviews = query(collection(db, 'reviews'), where('mediaId', '==', mediaId));
  return onSnapshot(mediaReviews, (snapshot) => {
    onReviews(snapshot.docs.map((entry) => ({ ...entry.data(), id: entry.id })));
  }, onError);
}

export async function addReview(userId, mediaId, reviewData) {
  const id = encodeURIComponent(JSON.stringify([userId, mediaId]));
  const entry = doc(db, 'reviews', id);
  await setDoc(entry, { ...reviewData, userId, mediaId });
  return entry;
}

export async function updateReview(reviewId, reviewData) {
  return updateDoc(doc(db, 'reviews', reviewId), reviewData);
}

export async function deleteReview(reviewId) {
  return deleteDoc(doc(db, 'reviews', reviewId));
}

