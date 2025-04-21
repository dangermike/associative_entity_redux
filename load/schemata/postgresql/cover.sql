DROP SCHEMA IF EXISTS cover CASCADE;
CREATE SCHEMA cover;

CREATE TABLE cover.people (
    id INT NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE INDEX IX_PEOPLE_NAME ON cover.people (name, id);

CREATE TABLE cover.companies (
    id INT NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE INDEX IX_COMPANIES_NAME ON cover.companies (name, id);

CREATE TABLE cover.people_companies (
    person_id INT NOT NULL,
    company_id INT NOT NULL,
    CONSTRAINT PK_PEOPLE_COMPANIES PRIMARY KEY (person_id, company_id)
);

CREATE INDEX IX_PEOPLE_COMPANIES ON cover.people_companies (company_id, person_id);