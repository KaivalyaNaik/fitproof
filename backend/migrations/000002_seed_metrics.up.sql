INSERT INTO metrics (name, unit, description) VALUES
    ('Steps',            'steps',   'Total steps walked or run in a day'),
    ('Water Intake',     'ml',      'Total water consumed in millilitres'),
    ('Sleep Duration',   'hours',   'Total hours of sleep'),
    ('Workout Duration', 'minutes', 'Total active exercise minutes'),
    ('Calories Burned',  'kcal',    'Estimated calories burned'),
    ('Pushups',          'reps',    'Number of pushups completed'),
    ('Running Distance', 'km',      'Distance run in kilometres'),
    ('Screen Time',      'minutes', 'Total screen time (use as a max metric)')
ON CONFLICT DO NOTHING;
