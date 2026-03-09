import React from 'react';

const ReviewSection = ({ reviews }) => {
    return (
        <div>
            <h2>User Reviews</h2>
            {reviews.length === 0 ? (
                <p>No reviews yet.</p>
            ) : (
                reviews.map(({ id, username, comment, rating, timestamp }) => (
                    <div key={id} className="review">
                        <h3>{username} - Rating: {rating}</h3>
                        <p>{comment}</p>
                        <small>{new Date(timestamp).toLocaleString()}</small>
                    </div>
                ))
            )}
        </div>
    );
};

export default ReviewSection;