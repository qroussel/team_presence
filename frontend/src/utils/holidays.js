import { addDays, getYear, isSameDay } from 'date-fns';

/**
 * Calculates Easter Sunday for a given year using the anonymous algorithm (Meeus/Jones/Butcher).
 * @param {number} year 
 * @returns {Date} Easter Sunday
 */
function getEasterSunday(year) {
    const a = year % 19;
    const b = Math.floor(year / 100);
    const c = year % 100;
    const d = Math.floor(b / 4);
    const e = b % 4;
    const f = Math.floor((b + 8) / 25);
    const g = Math.floor((b - f + 1) / 3);
    const h = (19 * a + b - d - g + 15) % 30;
    const i = Math.floor(c / 4);
    const k = c % 4;
    const l = (32 + 2 * e + 2 * i - h - k) % 7;
    const m = Math.floor((a + 11 * h + 22 * l) / 451);
    const month = Math.floor((h + l - 7 * m + 114) / 31);
    const day = ((h + l - 7 * m + 114) % 31) + 1;

    // Month is 1-indexed in algorithm, Date expects 0-indexed month
    return new Date(year, month - 1, day);
}

/**
 * Returns an array of French holidays for a given year.
 * @param {number} year 
 * @returns {Date[]} Array of holiday dates
 */
export const getFrenchHolidays = (year) => {
    const fixedHolidays = [
        new Date(year, 0, 1),   // New Year's Day
        new Date(year, 4, 1),   // Labor Day
        new Date(year, 4, 8),   // Victory Day
        new Date(year, 6, 14),  // Bastille Day
        new Date(year, 7, 15),  // Assumption of Mary
        new Date(year, 10, 1),  // All Saints' Day
        new Date(year, 10, 11), // Armistice Day
        new Date(year, 11, 25), // Christmas Day
    ];

    const easterSunday = getEasterSunday(year);
    const easterMonday = addDays(easterSunday, 1);
    const ascensionDay = addDays(easterSunday, 39);
    const whitMonday = addDays(easterSunday, 50);

    return [
        ...fixedHolidays,
        easterMonday,
        ascensionDay,
        whitMonday
    ];
};

/**
 * Checks if a given date is a French holiday.
 * @param {Date} date 
 * @returns {boolean}
 */
export const isFrenchHoliday = (date) => {
    if (!date) return false;
    const year = getYear(date);
    const holidays = getFrenchHolidays(year);
    return holidays.some(h => isSameDay(h, date));
};
