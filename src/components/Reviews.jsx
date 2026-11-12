import React, { useEffect, useState } from 'react';
import { addReview, deleteReview, subscribeToReviews, updateReview } from '../services/firestore';

export function Reviews({ mediaId, userId }) {
  const [reviews, setReviews] = useState(null);
  const [loadError, setLoadError] = useState('');
  const [error, setError] = useState('');
  const [draft, setDraft] = useState(null);
  const [pending, setPending] = useState(false);

  useEffect(() => subscribeToReviews(mediaId, setReviews, () => {
    setLoadError('Could not load reviews.');
  }), [mediaId]);

  async function submit(event) {
    event.preventDefault();
    const data = { rating: Number(draft.rating), text: draft.text.trim() };
    setPending(true);
    setError('');
    try {
      if (draft.id) await updateReview(draft.id, data);
      else await addReview(userId, mediaId, data);
      setDraft(null);
    } catch {
      setError('Could not save your review. Try again.');
    } finally {
      setPending(false);
    }
  }

  async function remove(id) {
    setPending(true);
    setError('');
    try {
      await deleteReview(id);
      setDraft(null);
    } catch {
      setError('Could not delete your review. Try again.');
    } finally {
      setPending(false);
    }
  }

  if (loadError) return <p role="alert">{loadError}</p>;
  if (!reviews) return <p role="status">Loading reviews…</p>;

  return (
    <div className="review-panel">
      {reviews.length === 0 && <p>No reviews yet.</p>}
      {reviews.map((review) => (
        <div key={review.id} className="review-entry">
          <strong>{review.rating}/5</strong>
          {review.text && <p>{review.text}</p>}
          {review.userId === userId && (
            <div className="collection-actions">
              <button type="button" disabled={pending} onClick={() => {
                setError('');
                setDraft({ id: review.id, rating: review.rating || 5, text: review.text || '' });
              }}>Edit review</button>
              <button type="button" disabled={pending} onClick={() => remove(review.id)}>Delete review</button>
            </div>
          )}
        </div>
      ))}
      {!draft && !reviews.some((review) => review.userId === userId) && (
        <button type="button" disabled={pending} onClick={() => {
          setError('');
          setDraft({ rating: 5, text: '' });
        }}>Write a review</button>
      )}
      {draft && (
        <form onSubmit={submit}>
          <label>
            Rating
            <select value={draft.rating} disabled={pending}
              onChange={(event) => setDraft({ ...draft, rating: event.target.value })}>
              {[1, 2, 3, 4, 5].map((rating) => <option key={rating}>{rating}</option>)}
            </select>
          </label>
          <label>
            Review
            <textarea value={draft.text} disabled={pending} maxLength={2000}
              onChange={(event) => setDraft({ ...draft, text: event.target.value })} />
          </label>
          <button type="submit" disabled={pending}>Save review</button>
          <button type="button" disabled={pending} onClick={() => {
            setDraft(null);
            setError('');
          }}>Cancel</button>
        </form>
      )}
      {error && <p role="alert">{error}</p>}
    </div>
  );
}
