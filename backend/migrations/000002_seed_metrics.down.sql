DELETE FROM metrics WHERE name IN (
    'Steps', 'Water Intake', 'Sleep Duration', 'Workout Duration',
    'Calories Burned', 'Pushups', 'Running Distance', 'Screen Time'
);
