create table if not exists timesheet_activity(
    timesheet_id ulid not null,
    activity_id integer not null,
    primary key (timesheet_id, activity_id),
    foreign key (timesheet_id) references timesheet(id) on delete cascade,
    foreign key (activity_id) references activity(id) on delete cascade
);