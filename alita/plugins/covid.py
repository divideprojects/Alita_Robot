import requests
import json
from dataclasses import dataclass
from typing import Any, List, TypeVar, Type, cast, Callable
from datetime import datetime
import dateutil.parser
from alita.__main__ import Alita
from pyrogram import filters
from pyrogram.types import Message
from alita import PREFIX_HANDLER

__PLUGIN__ = "Covid"

__help__ = """
Get Latest Corona Virus Stats in your group using this bot!

**Usage:**
 - /covid: Get Worldwide Corona Status
"""

T = TypeVar("T")


def from_str(x: Any) -> str:
    assert isinstance(x, str)
    return x


def from_int(x: Any) -> int:
    assert isinstance(x, int) and not isinstance(x, bool)
    return x


def from_datetime(x: Any) -> datetime:
    return dateutil.parser.parse(x)


def to_class(c: Type[T], x: Any) -> dict:
    assert isinstance(x, c)
    return cast(Any, x).to_dict()


def from_list(f: Callable[[Any], T], x: Any) -> List[T]:
    assert isinstance(x, list)
    return [f(y) for y in x]


@dataclass
class Premium:
    pass

    @staticmethod
    def from_dict(obj: Any) -> "Premium":
        assert isinstance(obj, dict)
        return Premium()

    def to_dict(self) -> dict:
        result: dict = {}
        return result


@dataclass
class Country:
    country: str
    country_code: str
    slug: str
    new_confirmed: int
    total_confirmed: int
    new_deaths: int
    total_deaths: int
    new_recovered: int
    total_recovered: int
    date: datetime
    premium: Premium

    @staticmethod
    def from_dict(obj: Any) -> "Country":
        assert isinstance(obj, dict)
        country = from_str(obj.get("Country"))
        country_code = from_str(obj.get("CountryCode"))
        slug = from_str(obj.get("Slug"))
        new_confirmed = from_int(obj.get("NewConfirmed"))
        total_confirmed = from_int(obj.get("TotalConfirmed"))
        new_deaths = from_int(obj.get("NewDeaths"))
        total_deaths = from_int(obj.get("TotalDeaths"))
        new_recovered = from_int(obj.get("NewRecovered"))
        total_recovered = from_int(obj.get("TotalRecovered"))
        date = from_datetime(obj.get("Date"))
        premium = Premium.from_dict(obj.get("Premium"))
        return Country(
            country,
            country_code,
            slug,
            new_confirmed,
            total_confirmed,
            new_deaths,
            total_deaths,
            new_recovered,
            total_recovered,
            date,
            premium,
        )

    def to_dict(self) -> dict:
        result: dict = {}
        result["Country"] = from_str(self.country)
        result["CountryCode"] = from_str(self.country_code)
        result["Slug"] = from_str(self.slug)
        result["NewConfirmed"] = from_int(self.new_confirmed)
        result["TotalConfirmed"] = from_int(self.total_confirmed)
        result["NewDeaths"] = from_int(self.new_deaths)
        result["TotalDeaths"] = from_int(self.total_deaths)
        result["NewRecovered"] = from_int(self.new_recovered)
        result["TotalRecovered"] = from_int(self.total_recovered)
        result["Date"] = self.date.isoformat()
        result["Premium"] = to_class(Premium, self.premium)
        return result


@dataclass
class Global:
    new_confirmed: int
    total_confirmed: int
    new_deaths: int
    total_deaths: int
    new_recovered: int
    total_recovered: int

    @staticmethod
    def from_dict(obj: Any) -> "Global":
        assert isinstance(obj, dict)
        new_confirmed = from_int(obj.get("NewConfirmed"))
        total_confirmed = from_int(obj.get("TotalConfirmed"))
        new_deaths = from_int(obj.get("NewDeaths"))
        total_deaths = from_int(obj.get("TotalDeaths"))
        new_recovered = from_int(obj.get("NewRecovered"))
        total_recovered = from_int(obj.get("TotalRecovered"))
        return Global(
            new_confirmed,
            total_confirmed,
            new_deaths,
            total_deaths,
            new_recovered,
            total_recovered,
        )

    def to_dict(self) -> dict:
        result: dict = {}
        result["NewConfirmed"] = from_int(self.new_confirmed)
        result["TotalConfirmed"] = from_int(self.total_confirmed)
        result["NewDeaths"] = from_int(self.new_deaths)
        result["TotalDeaths"] = from_int(self.total_deaths)
        result["NewRecovered"] = from_int(self.new_recovered)
        result["TotalRecovered"] = from_int(self.total_recovered)
        return result


@dataclass
class CoviData:
    message: str
    covi_data_global: Global
    countries: List[Country]
    date: datetime

    @staticmethod
    def from_dict(obj: Any) -> "CoviData":
        assert isinstance(obj, dict)
        message = from_str(obj.get("Message"))
        covi_data_global = Global.from_dict(obj.get("Global"))
        countries = from_list(Country.from_dict, obj.get("Countries"))
        date = from_datetime(obj.get("Date"))
        return CoviData(message, covi_data_global, countries, date)

    def to_dict(self) -> dict:
        result: dict = {}
        result["Message"] = from_str(self.message)
        result["Global"] = to_class(Global, self.covi_data_global)
        result["Countries"] = from_list(lambda x: to_class(Country, x), self.countries)
        result["Date"] = self.date.isoformat()
        return result


def covi_data_from_dict(s: Any) -> CoviData:
    return CoviData.from_dict(s)


def covi_data_to_dict(x: CoviData) -> Any:
    return to_class(CoviData, x)


@Alita.on_message(filters.command("covid", PREFIX_HANDLER))
async def get_covid(c: Alita, m: Message):

    cm = await m.reply_text(
        "**__Fetching stats...__**", reply_to_message_id=m.message_id
    )
    fetch = requests.get("https://api.covid19api.com/summary")
    result = covi_data_from_dict(json.loads(fetch.text))

    if fetch.status_code == 200:
        # Todays Data
        new_confirmed_global = result.covi_data_global.new_confirmed
        new_deaths_global = result.covi_data_global.new_deaths
        new_recovered_global = result.covi_data_global.new_recovered
        # All Time Data
        total_confirmed_global = result.covi_data_global.total_confirmed
        total_deaths_global = result.covi_data_global.total_deaths
        total_recovered_global = result.covi_data_global.total_recovered
        active_cases_covid19 = (
            total_confirmed_global - total_deaths_global - total_recovered_global
        )
        reply_text = (
            "**Corona Stats🦠:**\n"
            "**New**\n"
            f"New Confirmed: `{str(new_confirmed_global)}`\n"
            f"New Deaths: `{str(new_deaths_global)}`\n"
            f"New Recovered: `{str(new_recovered_global)}`\n"
            "\n**Total**\n"
            f"Total Confirmed: `{str(total_confirmed_global)}`\n"
            f"Total Deaths: `{str(total_deaths_global)}`\n"
            f"Total Recovered: `{str(total_recovered_global)}`\n"
            f"Active Cases: `{str(active_cases_covid19)}`"
        )

    await cm.edit_text(reply_text)
    return
