DROP SCHEMA IF EXISTS nocover CASCADE;
CREATE SCHEMA nocover;

CREATE TABLE nocover.people (
    id INT NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE INDEX IX_PEOPLE_NAME ON nocover.people (name);

CREATE TABLE nocover.companies (
    id INT NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE INDEX IX_COMPANIES_NAME ON nocover.companies (name);

CREATE TABLE nocover.people_companies (
    person_id INT NOT NULL,
    company_id INT NOT NULL,
    CONSTRAINT PK_PEOPLE_COMPANIES PRIMARY KEY (person_id, company_id)
);

CREATE INDEX IX_PEOPLE_COMPANIES ON nocover.people_companies (company_id);